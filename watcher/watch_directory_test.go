package watcher_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sevigo/notify"
	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/event"
	"github.com/sevigo/notify/watcher"
	"github.com/stretchr/testify/assert"
)

var directoryWatcher core.DirectoryWatcher

func init() {
	directoryWatcher = notify.Setup(context.TODO(), nil)
	go func() {
		for err := range directoryWatcher.Error() {
			fmt.Printf("[%s] %q\n", err.Level, err.Message)

		}
	}()
}

func TestStartWatching(t *testing.T) {
	watchPath := "testdata/"
	options := &core.WatchingOptions{
		Rescan:        true,
		IgnoreFiles:   []string{".html"},
		IgnoreFolders: []string{"testdata/ignore"},
	}
	go directoryWatcher.StartWatching(watchPath, options)

	var wg sync.WaitGroup
	wg.Add(1)
	var event event.Event
	go func() {
		event = <-directoryWatcher.Event()
		wg.Done()
	}()
	wg.Wait()

	assert.Equal(t, "added", watcher.ActionToString(event.Action))
	assert.Equal(t, "testdata/test.txt", event.Path)
	directoryWatcher.StopWatching(watchPath)
}

func TestRescan(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	var event event.Event
	go func() {
		event = <-directoryWatcher.Event()
		wg.Done()
	}()

	watchPath := "testdata/"
	options := &core.WatchingOptions{
		Rescan:        false,
		IgnoreFiles:   []string{".html"},
		IgnoreFolders: []string{"testdata/ignore"},
	}
	go directoryWatcher.StartWatching(watchPath, options)
	// give StartWatching some time to do the initial work
	time.Sleep(time.Second)
	directoryWatcher.RescanAll()

	wg.Wait()
	assert.Equal(t, "added", watcher.ActionToString(event.Action))
	assert.Equal(t, "testdata/test.txt", event.Path)
	directoryWatcher.StopWatching(watchPath)
}
