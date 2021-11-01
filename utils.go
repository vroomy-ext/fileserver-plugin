package plugin

import (
	"fmt"
	"path"
	"path/filepath"
)

func parseArgs(args []string) (dir, filename, pathRoot string, err error) {
	switch len(args) {
	case 1:
	case 2:
		pathRoot = args[1]

	default:
		err = fmt.Errorf("invalid number of arguments, expected 1-2 and received %d", len(args))
		return
	}

	if filepath.Ext(args[0]) == "" {
		dir = path.Clean(args[0])
		return
	}

	dir = path.Dir(args[0])
	filename = path.Base(args[0])
	return
}
