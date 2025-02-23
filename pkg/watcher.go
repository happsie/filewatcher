package pkg

import (
	"context"
	"log/slog"
	"runtime"
)

type OsWatcher interface {
	watch(dir string, eventChan chan WatchEvent, ctx context.Context) error
	unwatch() error
}

type handlerFunc func(event WatchEvent)

type Watcher struct {
	watcher  OsWatcher
	handlers map[ModificationType][]handlerFunc
	dir      string
	ctx      context.Context
}

// NewWatcher creates a watcher for the supplied dir (directory).
// NewWatcher will automatically determine the correct watcher based on operating system
func NewWatcher(dir string) *Watcher {
	var watcher OsWatcher
	if runtime.GOOS == "linux" {
		watcher = newLinuxFileWatcher()
	}
	return &Watcher{
		watcher:  watcher,
		handlers: make(map[ModificationType][]handlerFunc),
		dir:      dir,
	}

}

// Watch will start watching the directory and direct the events from the OS to handlers. This is a blocking call
// Cancelling the context will remove the watch and stop the event loop
func (w *Watcher) Watch(ctx context.Context) {
	w.ctx = ctx
	eventCh := make(chan WatchEvent)
	go func() {
		err := w.watcher.watch(w.dir, eventCh, ctx)
		if err != nil {
			slog.Error("error watching directory", "error", err)
		}
	}()

	select {
	case <-ctx.Done():
		return
	default:
		for event := range eventCh {
			handlers := w.handlers[event.ModificationType]
			for _, handler := range handlers {
				go handler(event)
			}
		}
	}
}

// Unwatch will remove the watcher from the watched directory
func (w *Watcher) Unwatch() error {
	w.ctx.Done()
	return w.watcher.unwatch()
}

// HandlerFunc will add a Handler function for the specified modificationType. Events from the operating system in the directory matching
// the modificationType will trigger the handler. The handler will consists of the event containing filename
func (w *Watcher) HandlerFunc(modificationType ModificationType, handler handlerFunc) {
	w.handlers[modificationType] = append(w.handlers[modificationType], handler)
}
