package flags

import "github.com/urfave/cli/v2"

var DatabaseFlags = []cli.Flag{
	SqliteFileFlag,
}

var Flags = append([]cli.Flag{
	DomainFilesFlag,
}, DatabaseFlags...)
