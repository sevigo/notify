package main

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/sevigo/notify"
	"github.com/sevigo/notify/watcher"
)

var dirsWin = []string{"C:\\Users\\Igor\\Files"}
var dirsLin = []string{"/home/igor/Downloads", "/home/igor/Documents"}

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

	w := notify.Setup(
		ctx,
		&watcher.Options{
			Rescan:      true,
			FileFilters: []string{".crdownload", ".lock", ".snapshot"},
		})

	for _, dir := range dirs {
		go w.StartWatching(dir)
	}
	defer func() {
		for _, dir := range dirs {
			w.StopWatching(dir)
		}
	}()

	fmt.Println("wait for file change events ...")
	// wait for file change events
	for {
		select {
		case ev := <-w.Event():
			log.Printf("[EVENT] %s: %q", watcher.ActionToString(ev.Action), ev.Path)
		case err := <-w.Error():
			if err.Level == "ERROR" {
				log.Printf("[%s] %s", err.Level, err.Message)
			}
		}
	}

}
