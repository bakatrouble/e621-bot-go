package worker

import (
	"cmp"
	"context"
	"e621-bot-go/e621"
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"maps"
	"slices"
	"strings"
	"time"
)

func CheckPostVersion(pv *e621.PostVersion, queries []*utils.QueryInfo) *utils.QueryInfo {
	currentTags := map[string]bool{}
	for _, tag := range strings.Split(pv.Tags, " ") {
		currentTags[tag] = true
	}

	prevTags := maps.Clone(currentTags)
	for _, tag := range pv.RemovedTags {
		prevTags[tag] = true
	}
	for _, tag := range pv.AddedTags {
		prevTags[tag] = false
	}

	for _, query := range queries {
		if query.Check(currentTags) && !query.Check(prevTags) {
			return query
		}
	}

	return nil
}

func checkPosts(ctx context.Context) error {
	logger := ctx.Value("logger").(utils.Logger)
	store := ctx.Value("store").(*storage.Storage)
	client := ctx.Value("e621").(*e621.E621)

	lock, err := store.Lock(ctx)
	if err != nil {
		logger.With("err", err).Error("failed to acquire lock")
		return err
	}
	defer func(ctx context.Context) {
		_ = lock.Release(ctx)
	}(ctx)

	queries, err := utils.GetQueries(store, ctx)
	if err != nil {
		logger.With("err", err).Error("failed to get queries")
		return err
	}

	lastPostVersion, err := store.GetLastPostVersion(ctx)
	if err != nil {
		logger.With("err", err).Error("failed to get last post version")
		return err
	}

	pvsToPost := make([]*e621.PostVersion, 0)
	const pageSize = 320
	for {
		rq := client.GetPostVersions().WithLimit(pageSize)
		if lastPostVersion != 0 {
			rq = rq.WithAfterID(lastPostVersion)
		}
		page, err := rq.Send(ctx)
		if err != nil {
			logger.With("err", err).Error("failed to fetch post versions")
			return err
		}
		if len(page) == 0 {
			break
		}
		slices.Reverse(page)
		for _, pv := range page {
			if pv.ID > lastPostVersion {
				lastPostVersion = pv.ID
			}

			//logger.With("post_version_id", pv.ID).With("post_id", pv.PostID).Debug("checking post version")
			if match := CheckPostVersion(pv, queries); match != nil {
				pvsToPost = append(pvsToPost, pv)
			}
		}
		logger.With("count", len(page)).Info("fetched post versions page")
		if len(page) < pageSize {
			break
		}
	}

	slices.SortFunc(pvsToPost, func(a, b *e621.PostVersion) int {
		return cmp.Compare(a.ID, b.ID)
	})

	postIds := make([]int, len(pvsToPost))
	for _, plan := range pvsToPost {
		postIds = append(postIds, plan.PostID)
	}

	sentFlags, err := store.IsPostSent(ctx, postIds)
	if err != nil {
		logger.With("err", err).Error("failed to check sent posts")
		return err
	}

	if len(pvsToPost) == 0 {
		logger.Info("no new posts to send")

		if err = store.SetLastPostVersion(ctx, lastPostVersion); err != nil {
			logger.With("err", err).Error("failed to update last post version")
		}

		return nil
	}

	ticker := time.NewTicker(3 * time.Second)
	for _, plan := range pvsToPost {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}

		if sentFlags[plan.PostID] {
			continue
		}
		if err = SendPost(ctx, client, plan.PostID, queries); err != nil {
			logger.With("err", err).With("post_version_id", plan.ID).Error("failed to send post")
			return err
		}
		if err = store.SetPostSent(ctx, plan.PostID); err != nil {
			logger.With("err", err).Error("failed to mark post as sent")
		}
		if err = store.SetLastPostVersion(ctx, plan.ID); err != nil {
			logger.With("err", err).Error("failed to update last post version")
		}
	}

	return nil
}
