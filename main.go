package main

import (
	"log/slog"
	"os"
)

func main() {
	watcher := NewWatcher(NewCachedDirectoryScanner())

	watcher.HandleFunc(Created, func(meta FileMeta) {
		slog.Info("handling created", "meta", meta)
	})
	watcher.HandleFunc(Modified, func(meta FileMeta) {
		slog.Info("handling modified", "meta", meta)
	})

	err := watcher.Watch()
	if err != nil {
		slog.Error("watcher failed", "error", err)
		os.Exit(1)
	}
}
