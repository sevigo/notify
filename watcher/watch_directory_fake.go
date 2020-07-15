// +build fake

package watcher

import (
	"time"

	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/event"
)

var ignoreFolders = map[string]bool{}

func (i *DirectoryWatcher) StartWatching(root string, _ *core.WatchingOptions) {
	time.Sleep(time.Second)
	fileChangeNotifier(root+"/test.txt", event.FileAdded, nil)
}
