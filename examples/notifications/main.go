package main

import (
	"context"
	"log"
	"runtime"
)

var dirsWin = []string{"C:\\Users\\Igor\\Files", "C:\\Users\\Igor\\Downloads"}
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

	w := watcher.Setup(
		ctx,
		&watcher.Options{
			IgnoreDirectoies: true,
			FileFilters:      []string{".crdownload", ".lock", ".snapshot"},
		})
	for _, dir := range dirs {
		w.StartWatching(dir)
	}
	defer func() {
		for _, dir := range dirs {
			w.StopWatching(dir)
		}
	}()
	for {
		select {
		case file := <-w.Events:
			log.Printf("[EVENT] %#v", file)
		case err := <-w.Errors:
			log.Printf("[%s] %s\n", err.Level, err.Message)
		}
	}
}
