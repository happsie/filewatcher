package main

import (
	"log/slog"
	"time"
)

type handlerFunc func(meta FileMeta)

type Watcher struct {
	scanner DirectoryScanner
	TimeoutMS int32
	handlers map[ModificationType]handlerFunc
}

func NewWatcher(scanner DirectoryScanner) Watcher {
	return Watcher{
		TimeoutMS: 1000,
		scanner: scanner,
		handlers: make(map[ModificationType]handlerFunc),
	} 
}

func (w *Watcher) Watch() error {
	for {
		events, err := w.scanner.Scan("test", ".")
		if err != nil {
			return err
		}
		for _, event := range events {
			slog.Info("event found", "event", event)
		}
		for _, event := range events {
			handler := w.handlers[event.Type]
			if handler == nil {
				slog.Debug("no handler found", "modification type", event.Type)
				continue
			}
			go handler(event)
		}
		time.Sleep(time.Duration(w.TimeoutMS) * time.Millisecond)
	}
}

func (w *Watcher) HandleFunc(modificationType ModificationType, handler func(meta FileMeta)) {
	w.handlers[modificationType] = handler	
}
