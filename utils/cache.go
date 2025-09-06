package utils

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func CacheFile(ctx context.Context, fileBytes []byte, name string) (string, error) {
	config := ctx.Value("config").(*Config)

	var err error
	path := filepath.Join(config.CacheDir, name)
	if err = os.WriteFile(path, fileBytes, os.ModePerm); err != nil {
		return "", err
	}

	return path, nil
}

func CacheCleaner(ctx context.Context) {
	config := ctx.Value("config").(*Config)
	logger := ctx.Value("logger").(Logger)

	var files []os.DirEntry
	var info fs.FileInfo
	var err error

	ticker := time.NewTicker(10 * time.Minute)
	for {
		select {
		case <-ticker.C:
			if files, err = os.ReadDir(config.CacheDir); err != nil {
				logger.With("err", err).Error("failed to read cache dir")
				continue
			}

			now := time.Now()
			for _, file := range files {
				info, err = file.Info()
				if err != nil {
					logger.With("err", err).Error("failed to get file info")
					continue
				}
				if now.Sub(info.ModTime()) > (10 * 24 * time.Hour) {
					if err = os.Remove(filepath.Join(config.CacheDir, file.Name())); err != nil {
						logger.With("err", err).With("file", file.Name()).Error("failed to remove cached file")
					} else {
						logger.With("file", file.Name()).Info("removed cached file")
					}
				}
			}
		case <-ctx.Done():
			logger.Info("stopping cache cleaner")
			return
		}
	}
}

func IsCached(ctx context.Context, name string) (string, bool) {
	config := ctx.Value("config").(*Config)

	path := filepath.Join(config.CacheDir, name)
	if _, err := os.Stat(path); err == nil {
		return path, true
	}
	return "", false
}
