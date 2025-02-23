# Filey

A small practice project exploring some deeper systems programming. Currently only working on linux.

## API 

```go
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
	watcher.Watch() // Blocking call, use go before to spin it up on a new go routine
}
```

## Run tests

```
make all-tests 
``` 
