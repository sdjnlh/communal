package util

import "unicode"

func LowerFirst(name string) string {
	var newName []rune
	for i, c := range name {
		if i == 0 {
			newName = append(newName, unicode.ToLower(c))
		} else {
			newName = append(newName, c)
		}
	}
	return string(newName)
}

func StringArrayContains(arr []string, str string) bool {
	for _, c := range arr {
		if str == c {
			return true
		}
	}
	return false
}
