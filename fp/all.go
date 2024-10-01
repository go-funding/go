package fp

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func MapErr[From, To any](fromSlice []From, fn func(From, int) (To, error)) ([]To, error) {
	return ReduceErr(fromSlice, func(from From, toSlice []To, fromIdx int) (_ []To, err error) {
		toSlice[fromIdx], err = fn(from, fromIdx)
		return toSlice, err
	}, make([]To, len(fromSlice)))
}

func ReduceErr[From, To any](fromSlice []From, fn func(From, To, int) (To, error), initialValue To) (to To, err error) {
	to = initialValue
	for fromIndex, from := range fromSlice {
		to, err = fn(from, to, fromIndex)
		if err != nil {
			return
		}
	}
	return
}

func GetFileFullPath(val string) (string, error) {
	if val[0] == '/' {
		return val, nil
	}

	ex, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return path.Join(ex, val), nil
}

func EnsureFileExists(file string) error {
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return errors.New(fmt.Sprintf("File %s does not exists", file))
	}
	return nil
}
