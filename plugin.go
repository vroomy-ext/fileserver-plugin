package plugin

import (
	"fmt"
	"log"
	"path"
	"path/filepath"

	"github.com/gdbu/fileserver"
	"github.com/gdbu/scribe"
	"github.com/hatchify/errors"
	"github.com/vroomy/common"
	"github.com/vroomy/vroomy"
)

var p Plugin

const (
	// ErrInvalidRoot is returned whe a root is longer than the request path
	ErrInvalidRoot = errors.Error("invalid root, cannot be longer than request path")
)

func init() {
	p.out = scribe.New("Fileserver")
	if err := vroomy.Register("fileserver", &p); err != nil {
		log.Fatal(err)
	}
}

type Plugin struct {
	vroomy.BasePlugin

	out *scribe.Scribe
}

// Methods to match vroomy.Plugin interface below

// Close will close the plugin
func (p *Plugin) Close() (err error) {
	return
}

// Handlers below

// ServeFile will serve a file
func (p *Plugin) ServeFile(args ...string) (h common.Handler, err error) {
	var (
		target string
		dir    string
		root   string

		isDir bool
	)

	switch len(args) {
	case 1:
		target = args[0]
	case 2:
		target = args[0]
		root = args[1]
		isDir = true
	default:
		err = fmt.Errorf("invalid number of arguments, expected 1-2 and received %d", len(args))
		return
	}

	target = path.Clean(target)
	dir = filepath.Dir(target)

	var fs *fileserver.FileServer
	if fs, err = fileserver.New(dir); err != nil {
		return
	}

	h = func(ctx common.Context) {
		var (
			key string
			err error
		)

		if isDir {
			if key, err = getKeyFromRequestPath(root, ctx.Request().URL.Path); err != nil {
				ctx.WriteString(400, "text/plain", err.Error())
				return
			}
		} else {
			key = filepath.Base(target)
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
