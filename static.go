package goyave

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5/util/fsutil"
)

func staticHandler(fs fs.StatFS, download bool) Handler {
	return func(response *Response, r *Request) {
		file := r.RouteParams["resource"]
		if !checkStaticPath(file) {
			response.Status(http.StatusNotFound)
			return
		}
		path := cleanStaticPath(fs, file)

		if download {
			response.Download(fs, path, path[lo.Clamp(strings.LastIndex(file, "/"), 0, len(path)):])
			return
		}
		response.File(fs, path)
	}
}

func cleanStaticPath(fs fs.StatFS, file string) string {
	file = strings.TrimPrefix(file, "/")
	path := file
	if path == "" {
		return "index.html"
	}
	if fsutil.IsDirectory(fs, strings.TrimSuffix(path, "/")) {
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		path += "index.html"
	}
	return filepath.Clean(path)
}

func checkStaticPath(path string) bool {
	if !strings.HasPrefix(path, "/") && len(path) > 0 { // Force leading slash if the path is not empty
		return false
	}
	if strings.Contains(path, "\\") || strings.Contains(path, "//") {
		return false
	}
	for _, ent := range strings.FieldsFunc(path, isSlash) {
		if ent == "." || ent == ".." || ent == "" {
			return false
		}
	}
	return true
}

func isSlash(r rune) bool {
	return r == '/'
}
