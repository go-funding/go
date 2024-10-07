package file_spawners

import (
	"errors"
	"fuk-funding/go/engine/application"
	"fuk-funding/go/utils/ufiles"
	"fuk-funding/go/utils/uurls"
)

func SpawnTxtFileDomains(fullFilePath string, app *application.App) {
	err := ufiles.IterateFileBySeparator(fullFilePath, []byte("\n"), func(domainBytes []byte) error {
		parsedDomain, err := uurls.ParseHost(string(domainBytes))
		if errors.Is(err, uurls.EmptyStringError) {
			return nil
		}

		if err != nil {
			return err
		}

		app.DomainSpawned.Publish(parsedDomain)
		return nil
	})

	if err != nil {
		return
	}
}
