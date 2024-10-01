package fp

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func Map[From, To any](ts []From, fn func(From) To) []To {
	result := make([]To, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func MapErr[From, To any](ts []From, fn func(From) (To, error)) ([]To, error) {
	result := make([]To, len(ts))
	for i, t := range ts {
		val, err := fn(t)
		result[i] = val
		if err != nil {
			return result, err
		}
	}

	return result, nil
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
