package kiva

import (
	"errors"
	"strings"
)

func ParseKey(txt string) (table, key string, err error) {
	if txt == "" {
		err = errors.New("key can't be blank")
		return
	}

	keys := strings.Split(txt, ":")
	if len(keys) < 2 {
		return "", "", errors.New("key format is table:id")
	}
	if keys[1] == "" {
		err = errors.New("key can't be blank")
		return
	}
	table = keys[0]
	key = keys[1]
	return
}

func Iif(logic bool, resTrue, resFalse interface{}) interface{} {
	if logic {
		return resTrue
	}
	return resFalse
}
