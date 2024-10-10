package file_spawners

import (
	"errors"
	"fuk-funding/go/utils/uurls"
	"github.com/mgorunuch/gosuper"
	"os"
)

func RunTxtFileBytes(fullFilePath string, react func([]byte)) {
	file, err := os.Open(fullFilePath)
	if err != nil {
		return
	}
	defer file.Close()

	iter := gosuper.NewReaderSeparatedIterator(file, []byte("\n"))
	for iter.Next() {
		var bytes []byte
		err := iter.Scan(&bytes)
		if err != nil {
			return
		}

		react(bytes)
	}
}

func TxtFileDomainCallback(fullFilePath string, callback func(string)) {
	RunTxtFileBytes(fullFilePath, func(bytes []byte) {
		parsedDomain, err := uurls.ParseHost(string(bytes))
		if errors.Is(err, uurls.EmptyStringError) {
			return
		}

		if err != nil {
			return
		}

		callback(parsedDomain)
	})
}
