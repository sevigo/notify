package util

import (
	"crypto/sha512"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
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
