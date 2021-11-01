package plugin

import (
	"fmt"
	"log"
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

// Handlers below

// ServeFile will serve a file
func (p *Plugin) ServeFile(args ...string) (h common.Handler, err error) {
	var (
		dir      string
		filename string
		pathRoot string
	)

	if dir, filename, pathRoot, err = parseArgs(args); err != nil {
		return
	}

	var fs *fileserver.FileServer
	if fs, err = fileserver.New(dir); err != nil {
		return
	}

	var getTarget func(ctx common.Context) (target string, err error)
	if len(filename) == 0 {
		getTarget = func(ctx common.Context) (target string, err error) {
			return getKeyFromRequestPath(pathRoot, ctx.Request().URL.Path)
		}
	} else {
		getTarget = func(ctx common.Context) (target string, err error) {
			return filename, nil
		}
	}

	h = func(ctx common.Context) {
		var (
			target string
			err    error
		)

		if target, err = getTarget(ctx); err != nil {
			ctx.WriteString(400, "text/plain", err.Error())
			return
		}

		if err := fs.Serve(target, ctx.Writer(), ctx.Request()); err != nil {
			err = fmt.Errorf("Error serving %s: %v", target, err)
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
