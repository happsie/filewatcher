package main

import (
	"log/slog"

	"github.com/happsie/filey/pkg"
)

func main() {
	watcher := pkg.NewWatcher("./test")
	watcher.HandlerFunc(pkg.Modified, func(event pkg.WatchEvent) {
		slog.Info("modified - handler 1", "event", event)
	})
	watcher.HandlerFunc(pkg.Modified, func(event pkg.WatchEvent) {
		slog.Info("modified - handler 2", "event", event)
	})
	watcher.HandlerFunc(pkg.Created, func(event pkg.WatchEvent) {
		slog.Info("created - handler 1", "event", event)
	})
	watcher.Start()
}
