package utils

import (
	"encoding/json"
	"math/rand"
	"regexp"
	"unicode/utf8"
)

var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func RemoveNotValidUtf8InString(s string) string {
	ret := s
	if !utf8.ValidString(s) {
		v := make([]rune, 0, len(s))
		for i, r := range s {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(s[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, r)
		}
		ret = string(v)
	}
	return ret
}

func RemoveANSI(input string) string {
	return ansiEscapeRegex.ReplaceAllString(input, "")
}

func RandString(n int) (ret string) {
	allString := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM0123456789"
	ret = ""
	for range n {
		r := rand.Intn(len(allString))
		ret = ret + allString[r:r+1]
	}
	return
}

func JsonStrToStruct[T any](str string) T {
	var data T
	json.Unmarshal([]byte(str), &data)
	return data
}

func StructToJsonStr[T any](data T) string {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonStr)
}
