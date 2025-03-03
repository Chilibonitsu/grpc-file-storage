package utils

import (
	"mime"
	"path/filepath"
	"strings"
)

func CheckFileName(fileName string) bool {
	if fileName == "" {
		return false
	}

	forbiddenChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range forbiddenChars {
		if strings.Contains(fileName, char) {
			return false
		}
	}

	if strings.Contains(fileName, "..") || strings.Contains(fileName, "./") {
		return false
	}

	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		contentType = "application/octet-stream" // По умолчанию
	}

	// allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
	// if !contains(allowedTypes, contentType) {
	//     return false
	// }

	return true
}

func GetExt(filename string) string {
	ext := filepath.Ext(filename)

	return ext
}
