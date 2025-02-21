package pkg

import (
	"log/slog"
	"runtime"
)

type OsWatcher interface {
	watch(dir string, eventChan chan WatchEvent) error
	unwatch() error
}

type handlerFunc func(event WatchEvent)

type Watcher struct {
	watcher  OsWatcher
	handlers map[ModificationType][]handlerFunc
	dir      string
}

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

func (w *Watcher) Watch() {
	eventCh := make(chan WatchEvent)
	go func() {
		err := w.watcher.watch(w.dir, eventCh)
		if err != nil {
			slog.Error("error watching directory", "error", err)
		}
	}()

	for event := range eventCh {
		handlers := w.handlers[event.ModificationType]
		for _, handler := range handlers {
			go handler(event)
		}
	}
}

func (w *Watcher) Unwatch() error {
	return w.watcher.unwatch()
}

func (w *Watcher) HandlerFunc(modificationType ModificationType, handler handlerFunc) {
	w.handlers[modificationType] = append(w.handlers[modificationType], handler)
}
