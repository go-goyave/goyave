package validation

import (
	"strconv"
	"strings"

	"goyave.dev/goyave/v4/util/fsutil"
	"goyave.dev/goyave/v4/util/sliceutil"
)

func validateFile(ctx *Context) bool {
	_, ok := ctx.Value.([]fsutil.File)
	return ok
}

func validateMIME(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	if ok {
		for _, file := range files {
			mime := file.MIMEType
			if i := strings.Index(mime, ";"); i != -1 { // Ignore MIME settings (example: "text/plain; charset=utf-8")
				mime = mime[:i]
			}
			if !sliceutil.ContainsStr(ctx.Rule.Params, mime) {
				return false
			}
		}
		return true
	}
	return false
}

func validateImage(ctx *Context) bool {
	ctx.Rule.Params = []string{"image/jpeg", "image/png", "image/gif", "image/bmp", "image/svg+xml", "image/webp"}
	return validateMIME(ctx)
}

func validateExtension(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	if ok {
		for _, file := range files {
			if i := strings.LastIndex(file.Header.Filename, "."); i != -1 {
				if !sliceutil.ContainsStr(ctx.Rule.Params, file.Header.Filename[i+1:]) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	}
	return false
}

func validateCount(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	size, err := strconv.Atoi(ctx.Rule.Params[0])
	if err != nil {
		panic(err)
	}

	if ok {
		return len(files) == size
	}

	return false
}

func validateCountMin(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	size, err := strconv.Atoi(ctx.Rule.Params[0])
	if err != nil {
		panic(err)
	}

	if ok {
		return len(files) >= size
	}

	return false
}

func validateCountMax(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	size, err := strconv.Atoi(ctx.Rule.Params[0])
	if err != nil {
		panic(err)
	}

	if ok {
		return len(files) <= size
	}

	return false
}

func validateCountBetween(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	min, errMin := strconv.Atoi(ctx.Rule.Params[0])
	max, errMax := strconv.Atoi(ctx.Rule.Params[1])
	if errMin != nil {
		panic(errMin)
	}
	if errMax != nil {
		panic(errMax)
	}

	if ok {
		length := len(files)
		return length >= min && length <= max
	}

	return false
}
