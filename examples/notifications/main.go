package main

import (
	"context"
	"log"
	"runtime"

	"github.com/sevigo/notify"
	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/watcher"
)

var (
	dirsWin = []string{"C:\\Users\\Igor\\Files\\test"}
	dirsLin = []string{"/home/igor"}
	dirsMac = []string{"/Users/igor.komlew/Downloads/xxx"}
)

func main() {
	log.Println("Starting the service ...")
	ctx := context.TODO()

	var dirs []string
	switch runtime.GOOS {
	case "windows":
		dirs = dirsWin
	case "linux":
		dirs = dirsLin
	case "darwin":
		dirs = dirsMac
	default:
		panic("not supported OS: " + runtime.GOOS)
	}

	w := notify.Setup(ctx, &watcher.Options{})

	for _, dir := range dirs {
		go w.StartWatching(dir, &core.WatchingOptions{
			Rescan: true,
		})
	}
	defer func() {
		for _, dir := range dirs {
			w.StopWatching(dir)
		}
	}()

	log.Println("wait for file change events ...")
	for {
		select {
		case ev := <-w.Event():
			log.Printf("[EVENT] %s: %q", watcher.ActionToString(ev.Action), ev.Path)
			log.Printf(">>> %+v\n", ev.AdditionalInfo)
		case err := <-w.Error():
			if err.Level == "CRITICAL" {
				log.Printf("[%s] %s", err.Level, err.Message)
			}
		}
	}
}
