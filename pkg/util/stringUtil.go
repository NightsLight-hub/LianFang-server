package util

import (
	"strconv"
	"strings"
)

func SplitBySlash(str string) []string {
	temp := strings.Split(str, "/")
	result := make([]string, 0)
	for _, item := range temp {
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func StrToInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}
