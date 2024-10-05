package ufiles

import "os"

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
