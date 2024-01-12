package goyave

import (
	"io/fs"
	"strings"

	"goyave.dev/goyave/v5/util/fsutil"
)

func staticHandler(fs fs.StatFS, download bool) Handler {
	return func(response *Response, r *Request) {
		file := r.RouteParams["resource"]
		path := cleanStaticPath(fs, file)

		if download {
			response.Download(fs, path, file[strings.LastIndex(file, "/")+1:])
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
	return path
}
