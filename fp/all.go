package fp

import (
	"bufio"
	"errors"
	"fmt"
	"go.uber.org/multierr"
	"io"
	"os"
	"path"
	"strings"
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

func IterateFileBytes(filePath string, f func(r byte) error) (err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&err, multierr.Close(file))

	r := bufio.NewReader(file)
	for {
		byteVal, err := r.ReadByte()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		err = f(byteVal)
		if err != nil {
			return err
		}
	}

	return err
}

func IterateFileBySeparator(filePath string, sep []byte, f func(bt []byte) error) error {
	var buff []byte
	err := IterateFileBytes(filePath, func(r byte) error {
		buff = append(buff, r)

		if IsEndsWith(buff, sep) {
			defer func() { buff = []byte{} }()
			if len(buff) == 0 {
				return nil
			}

			buff = buff[0:SliceIdx(buff, -len(sep))]
			return f(buff)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if len(buff) > 0 {
		return f(buff)
	}

	return nil
}

func SliceIdx[Type any](v []Type, i int) int {
	l := len(v)
	return ((i % l) + l) % l
}

func SliceAt[Type any](v []Type, i int) Type {
	return v[SliceIdx(v, i)]
}

func IsEndsWith[Type comparable](source []Type, ends []Type) bool {
	if len(ends) > len(source) {
		return false
	}

	if len(ends) == 0 || len(source) == 0 {
		return false
	}

	for i := 0; i < len(ends); i++ {
		idx := -(i + 1)
		if SliceAt(ends, idx) != SliceAt(source, idx) {
			return false
		}
	}

	return true
}

func StrTrim(v string) string {
	return strings.Trim(v, "\n\t \r")
}

func IsStrEmpty(v string) bool {
	return v == ``
}
