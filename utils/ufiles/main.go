package ufiles

import (
	"bufio"
	"errors"
	"fmt"
	"fuk-funding/go/fp"
	"go.uber.org/multierr"
	"io"
	"os"
	"path"
)

func AppendToFile(path string, text string) (err error) {
	var f *os.File
	if f, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600); err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		return err
	}

	return nil
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

		if fp.IsEndsWith(buff, sep) {
			defer func() { buff = []byte{} }()
			if len(buff) == 0 {
				return nil
			}

			buff = buff[0:fp.SliceIdx(buff, -len(sep))]
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
