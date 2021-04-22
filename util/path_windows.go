package util

import (
	"os"
	"regexp"
)

func GetPath(path string) string {
	re := regexp.MustCompile(`(%([^/%]+)%)`)
	return re.ReplaceAllStringFunc(path, func(s string) string {
		return os.Getenv(regexp.MustCompile(`%([^/%]+)%`).FindStringSubmatch(s)[1])
	})
}
