package plugin

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gdbu/fileserver"
	"github.com/gdbu/scribe"
	"github.com/hatchify/errors"
	"github.com/vroomy/common"
	"github.com/vroomy/plugins"
)

var p Plugin

const (
	// ErrInvalidRoot is returned whe a root is longer than the request path
	ErrInvalidRoot = errors.Error("invalid root, cannot be longer than request path")
)

func init() {
	p.out = scribe.New("Fileserver")
	if err := plugins.Register("fileserver", &p); err != nil {
		log.Fatal(err)
	}
}

type Plugin struct {
	plugins.BasePlugin

	out *scribe.Scribe
}

// Methods to match plugins.Plugin interface below

// Close will close the plugin
func (p *Plugin) Close() (err error) {
	return
}

// Handlers below

// ServeFile will serve a file
func (p *Plugin) ServeFile(args ...string) (h common.Handler, err error) {
	var dir, root string
	if len(args) != 2 {
		err = fmt.Errorf("invalid number of arguments, expected %d and received %d", 2, len(args))
		return
	}

	dir = args[0]
	root = args[1]

	var fs *fileserver.FileServer
	if fs, err = fileserver.New(dir); err != nil {
		return
	}

	h = func(ctx common.Context) {
		var (
			key string
			err error
		)

		if key, err = getKeyFromRequestPath(root, ctx.Request().URL.Path); err != nil {
			ctx.WriteString(400, "text/plain", err.Error())
			return
		}

		if err := fs.Serve(key, ctx.Writer(), ctx.Request()); err != nil {
			err = fmt.Errorf("Error serving %s: %v", key, err)
			ctx.WriteString(400, "text/plain", err.Error())
			return
		}
	}

	return
}

func getKeyFromRequestPath(root, requestPath string) (key string, err error) {
	// Clean request path
	requestPath = filepath.Clean(requestPath)

	if len(root) > len(requestPath) {
		err = ErrInvalidRoot
		return
	}

	key = requestPath[len(root):]
	return
}
