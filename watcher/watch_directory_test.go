package watcher_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	watchPath := "testdata"
	options := &core.WatchingOptions{
		Rescan: true,
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

	expectedDir := filepath.Join("testdata", "test.txt")
	assert.Equal(t, "added", watcher.ActionToString(event.Action))
	assert.Equal(t, expectedDir, event.Path)
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

	watchPath := "testdata"
	options := &core.WatchingOptions{
		Rescan: false,
	}
	go directoryWatcher.StartWatching(watchPath, options)
	// give StartWatching some time to do the initial work
	time.Sleep(time.Second)
	directoryWatcher.RescanAll()

	wg.Wait()
	expectedDir := filepath.Join("testdata", "test.txt")
	assert.Equal(t, "added", watcher.ActionToString(event.Action))
	assert.Equal(t, expectedDir, event.Path)
	directoryWatcher.StopWatching(watchPath)
}

func TestNotification(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	var event event.Event
	go func() {
		event = <-directoryWatcher.Event()
		wg.Done()
	}()

	watchPath := "testdata"
	options := &core.WatchingOptions{
		Rescan: false,
	}
	go directoryWatcher.StartWatching(watchPath, options)
	// give StartWatching some time to do the initial work
	time.Sleep(time.Second)

	expectedDir := filepath.Join("testdata", "new.file")
	err := ioutil.WriteFile(expectedDir, []byte("Hello"), 0755)
	assert.NoError(t, err)

	wg.Wait()
	assert.Equal(t, "modified", watcher.ActionToString(event.Action))
	assert.Equal(t, expectedDir, event.Path)
	directoryWatcher.StopWatching(watchPath)

	err = os.Remove(expectedDir)
	assert.NoError(t, err)
}
