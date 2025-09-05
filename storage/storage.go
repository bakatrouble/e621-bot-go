package storage

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

const KeyPrefix = "e621-go"
const SubsKey = KeyPrefix + ":subs"
const SentKey = KeyPrefix + ":sent"
const LastPostVersionKey = KeyPrefix + ":last_post_version"
const LockKey = KeyPrefix + ":lock"

type Storage struct {
	client *redis.Client
	locker *redislock.Client
}

func NewStorage(url string) *Storage {
	opts, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(opts)
	locker := redislock.New(client)
	return &Storage{
		client,
		locker,
	}
}

func (s *Storage) GetSubs(ctx context.Context) ([]string, error) {
	r, err := s.client.SMembers(ctx, SubsKey).Result()
	if err != nil {
		return nil, err
	}
	slices.Sort(r)
	return r, nil
}

func (s *Storage) AddSub(ctx context.Context, sub string) error {
	if sub == "" {
		return nil
	}
	return s.client.SAdd(ctx, SubsKey, sub).Err()
}

func (s *Storage) RemoveSub(ctx context.Context, sub string) error {
	if sub == "" {
		return nil
	}
	return s.client.SRem(ctx, SubsKey, sub).Err()
}

func (s *Storage) IsPostSent(ctx context.Context, postIDs []int) (map[int]bool, error) {
	result := make(map[int]bool)
	if len(postIDs) == 0 {
		return result, nil
	}

	pids := make([]interface{}, len(postIDs))
	for i, id := range postIDs {
		pids[i] = id
	}

	val, err := s.client.SMIsMember(ctx, SentKey, pids...).Result()
	if !errors.Is(err, redis.Nil) {
		return nil, err
	}
	for i, id := range postIDs {
		result[id] = val[i]
	}
	return result, nil
}

func (s *Storage) SetPostSent(ctx context.Context, postID int) error {
	if err := s.client.SAdd(ctx, SentKey, postID).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) GetLastPostVersion(ctx context.Context) (int, error) {
	val, err := s.client.Get(ctx, LastPostVersionKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, err
	}
	return val, nil
}

func (s *Storage) SetLastPostVersion(ctx context.Context, version int) error {
	if err := s.client.Set(ctx, LastPostVersionKey, version, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) Lock(ctx context.Context) (*redislock.Lock, error) {
	return s.locker.Obtain(ctx, LockKey, 100*time.Millisecond, nil)
}
