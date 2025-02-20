package pkg

import (
	"log/slog"
	"runtime"
)

type IWatcher interface {
	watch(dir string, eventChan chan WatchEvent) error
}

type handlerFunc func(event WatchEvent)

type Watcher struct {
	watcher  IWatcher
	handlers map[ModificationType][]handlerFunc
	dir      string
}

func NewWatcher(dir string) *Watcher {
	var watcher IWatcher
	if runtime.GOOS == "linux" {
		watcher = newLinuxFileWatcher()
	}
	return &Watcher{
		watcher:  watcher,
		handlers: make(map[ModificationType][]handlerFunc),
		dir:      dir,
	}

}

func (w *Watcher) Start() {
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
			handler(event)
		}
	}
}

func (w *Watcher) HandlerFunc(modificationType ModificationType, handler handlerFunc) {
	w.handlers[modificationType] = append(w.handlers[modificationType], handler)
}
