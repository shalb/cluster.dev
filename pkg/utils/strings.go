package utils

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"regexp"
	"time"
)

// TruncateString - truncate string src to maxLen.
func TruncateString(src string, maxLen int) string {

	if len(src) <= maxLen {
		return src
	}

	res := src[0 : maxLen-1]
	return res
}

func Md5(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// URLToFolderName convert URL to string which can be used as a directory name.
func URLToFolderName(URL string) (string, error) {
	reg, err := regexp.Compile(`[^A-Za-z0-9_\-=\?\.]+`)
	if err != nil {
		return "", err
	}
	newStr := reg.ReplaceAllString(URL, "_")
	return newStr, nil
}
