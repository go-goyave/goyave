package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/util/fsutil"
)

type ParseRequest struct {
	// FIXME import path is not convenient because it clashes with custom middleware package in convention (move it to "goyave" package?)
	goyave.Controller
}

func (m *ParseRequest) Handle(next goyave.HandlerV5) goyave.HandlerV5 {
	return func(response *goyave.ResponseV5, r *goyave.RequestV5) {

		if err := parseQuery(r); err != nil {
			response.Status(http.StatusBadRequest)
			return
		}

		r.Data = nil
		contentType := r.Header().Get("Content-Type")
		if contentType != "" {
			maxSize := int64(m.Config().GetFloat("server.maxUploadSize") * 1024 * 1024)
			maxValueBytes := maxSize
			var bodyBuf bytes.Buffer
			n, err := io.CopyN(&bodyBuf, r.Request().Body, maxValueBytes+1)
			if err == nil || err == io.EOF {
				maxValueBytes -= n
				if maxValueBytes < 0 {
					response.Status(http.StatusRequestEntityTooLarge)
					return
				}

				bodyBytes := bodyBuf.Bytes()
				if strings.HasPrefix(contentType, "application/json") {
					var body any
					if err := json.Unmarshal(bodyBytes, &body); err != nil {
						response.Status(http.StatusBadRequest)
					}
					r.Data = body
				} else {
					r.Data, err = generateFlatMap(r.Request(), maxSize)
					if err != nil {
						response.Status(http.StatusBadRequest)
					}
				}
			} else {
				response.Status(http.StatusBadRequest)
			}
		}
		_ = r.Request().Body.Close()

		if response.GetStatus() != http.StatusBadRequest {
			next(response, r)
		}
	}
}

func parseQuery(request *goyave.RequestV5) error {
	queryParams, err := url.ParseQuery(request.URL().RawQuery)
	if err == nil {
		request.Query = make(map[string]any, len(queryParams))
		flatten(request.Query, queryParams)
	}
	return err
}

func generateFlatMap(request *http.Request, maxSize int64) (map[string]any, error) {
	flatMap := make(map[string]any)
	err := request.ParseMultipartForm(maxSize)

	if err != nil {
		if err == http.ErrNotMultipart {
			if err := request.ParseForm(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if request.Form != nil {
		flatten(flatMap, request.Form)
	}
	if request.MultipartForm != nil {
		flatten(flatMap, request.MultipartForm.Value)

		for field := range request.MultipartForm.File {
			flatMap[field] = fsutil.ParseMultipartFiles(request, field)
		}
	}

	// Source form is not needed anymore, clear it.
	request.Form = nil
	request.PostForm = nil
	request.MultipartForm = nil

	return flatMap, nil
}

func flatten(dst map[string]any, values url.Values) {
	for field, value := range values {
		if len(value) > 1 {
			dst[field] = value
		} else {
			dst[field] = value[0]
		}
	}
}
