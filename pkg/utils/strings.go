package utils

import (
	"crypto/md5"
	"fmt"
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
