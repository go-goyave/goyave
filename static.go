package goyave

import (
	"strings"

	"goyave.dev/goyave/v5/util/fsutil"
)

func staticHandler(directory string, download bool) Handler {
	return func(response *Response, r *Request) {
		file := r.RouteParams["resource"]
		path := cleanStaticPath(directory, file)

		if download {
			response.Download(path, file[strings.LastIndex(file, "/")+1:])
			return
		}
		response.File(path)
	}
}

func cleanStaticPath(directory string, file string) string {
	file = strings.TrimPrefix(file, "/")
	path := directory + "/" + file
	if fsutil.IsDirectory(path) {
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		path += "index.html"
	}
	return path
}
