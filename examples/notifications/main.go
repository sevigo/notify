package main

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/sevigo/notify"
	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/event"
	"github.com/sevigo/notify/watcher"
)

var dirsWin = []string{"C:\\Users\\Igor"}
var dirsLin = []string{"/home/igor"}

func main() {
	log.Println("Starting the service ...")
	ctx := context.TODO()

	var dirs []string
	switch runtime.GOOS {
	case "windows":
		dirs = dirsWin
	case "linux":
		dirs = dirsLin
	default:
		panic("not supported OS")
	}

	w := notify.Setup(ctx, &watcher.Options{})

	for _, dir := range dirs {
		go w.StartWatching(dir, &core.WatchingOptions{
			Rescan:        true,
			IgnoreFolders: []string{".vscode", ".atom", "AppData", "go", ".VirtualBox", ".git"},
			IgnoreFiles:   []string{".xml", ".json", ".ini"},
			AcceptFiles:   []string{".jpg"},
		})
	}
	defer func() {
		for _, dir := range dirs {
			w.StopWatching(dir)
		}
	}()

	log.Println("wait for file change events ...")
	total := 0
	var size int64
	for {
		select {
		case ev := <-w.Event():
			if ev.Action == event.FileAdded {
				total++
				if ev.Size > 0 {
					size = size + ev.Size
				}
				log.Printf("[EVENT] %s: %q", watcher.ActionToString(ev.Action), ev.Path)
			}
		case err := <-w.Error():
			if err.Level == "CRITICAL" {
				log.Printf("[%s] %s", err.Level, err.Message)
			}
		case <-time.After(3 * time.Second):
			log.Printf("found %d files\nsize=%d GB\n", total, size/1024/1024/1024)
			return
		}
	}
}
