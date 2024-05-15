package fileutil

import (
	"crypto/sha512"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"

	"github.com/sevigo/notify/event"
)

// ContentType returns mime type of the file as a string
// source: https://golangcode.com/get-the-content-type-of-file/
func ContentType(absoluteFilePath string) (string, error) {
	out, err := os.Open(path.Clean(absoluteFilePath))
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

// Checksum returns a string representation of SHA-512/256 checksum
func Checksum(absoluteFilePath string) (string, error) {
	f, err := os.Open(path.Clean(absoluteFilePath))
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha512.New512_256()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func CheckValidFile(absoluteFilePath string, action event.ActionType) (fs.FileInfo, error) {
	// if the file is removed we are good and the event is valid
	if action == event.FileRemoved {
		return nil, nil
	}
	if action == event.FileRenamedOldName {
		return nil, fmt.Errorf("FileRenamedOldName")
	}

	// we are checking this because windows tend to create some tmp files if this is a download files
	starts, err := os.Stat(absoluteFilePath)
	if err != nil {
		return nil, err
	}
	if starts.IsDir() {
		return nil, fmt.Errorf("not a file")
	}
	return starts, nil
}
