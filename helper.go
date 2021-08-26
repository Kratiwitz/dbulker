package main

import "strings"

func slugify(str string) string {
	str = strings.ReplaceAll(str, " ", "_")
	str = strings.ToLower(str)

	return str
}
