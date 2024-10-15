package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

func CheckAuth(r *http.Request, key string) bool {
	return r.Header.Get("X-Auth") == key
}

func FormatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
}

func HashFile(file multipart.File, handler *multipart.FileHeader) (string, error) {
	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	sha1Hash := hex.EncodeToString(hasher.Sum(nil))[:8]

	filename := fmt.Sprintf("%s%s", sha1Hash, filepath.Ext(handler.Filename))

	return filename, nil
}
