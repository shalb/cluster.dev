package utils

// TruncateString - truncate string src to maxLen.
func TruncateString(src string, maxLen int) string {

	if len(src) <= maxLen {
		return src
	}

	res := src[0 : maxLen-1]
	return res
}
