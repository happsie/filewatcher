package main

import (
	"log/slog"

	"github.com/happsie/filey/internal"
)

func main() {
	eventChan := make(chan internal.WatchEvent)
	watcher := internal.NewLinuxFileWatcher("./test")
	go func() {
		err := watcher.Watch(eventChan)
		if err != nil {
			slog.Error("error watching directory", "error", err)
		}
	}()

	for {
		select {
		case <-eventChan: 
			slog.Info("new event", "event", <-eventChan)
	}

	}
		/*
	watcher := NewWatcher(NewCachedDirectoryScanner())

	watcher.HandleFunc(Created, func(meta FileMeta) {
		slog.Info("handling created", "meta", meta)
	})
	watcher.HandleFunc(Modified, func(meta FileMeta) {
		slog.Info("handling modified", "meta", meta)
	})
	watcher.HandleFunc(Removed, func(meta FileMeta) {
		slog.Info("handling removed", "meta", meta)
	})

	err := watcher.Watch()
	if err != nil {
		slog.Error("watcher failed", "error", err)
		os.Exit(1)
	}*/
}
