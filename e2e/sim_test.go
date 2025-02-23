package e2e

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/happsie/filey/pkg"
)

type TestResult struct {
	mu       sync.Mutex
	created  int
	modified int
	deleted  int
}

func TestSimulator(t *testing.T) {
	result := TestResult{}

	watcher := pkg.NewWatcher("./test")
	watcher.HandlerFunc(pkg.Modified, func(event pkg.WatchEvent) {
		result.mu.Lock()
		result.modified++
		result.mu.Unlock()
	})
	watcher.HandlerFunc(pkg.Deleted, func(event pkg.WatchEvent) {
		result.mu.Lock()
		result.deleted++
		result.mu.Unlock()
	})
	watcher.HandlerFunc(pkg.Created, func(event pkg.WatchEvent) {
		result.mu.Lock()
		result.created++
		result.mu.Unlock()

	})
	go watcher.Watch()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 1; i <= 1000; i++ {
			fileName := fmt.Sprintf("./test/test_create_%d", i)
			os.Create(fileName)
			os.Remove(fileName)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for i := 1; i <= 1000; i++ {
			fileName := fmt.Sprintf("./test/test_modify_%d", i)
			os.WriteFile(fileName, []byte("hello"), os.ModePerm)
			os.Remove(fileName)
		}
		wg.Done()
	}()

	wg.Wait()
	// We are sleeping here for 5 seconds to wait for all events from the OS to come through
	time.Sleep(5 * time.Second)
	watcher.Unwatch()

	// Created is 2000 since we are also simultanously creating modified files. 1. Create, 2. Write "hello" to file
	if result.created != 2000 {
		t.Log("created files mismatch!", "files", result.created)
		t.Fail()
	}
	// Deleted is 2000, since we are also deleting our created files
	if result.deleted != 2000 {
		t.Log("deleted files mismatch!", "files", result.deleted)
		t.Fail()
	}
	if result.modified != 1000 {
		t.Log("modified files mismatch!", "files", result.modified)
		t.Fail()
	}
}
