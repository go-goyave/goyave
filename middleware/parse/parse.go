package parse

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

// TODO document and test parse middleware

type Middleware struct {
	goyave.Component

	// MaxUpoadSize the maximum size of the request (in MiB).
	// Defaults to the value provided in the config "server.maxUploadSize".
	MaxUploadSize float64
}

func (m *Middleware) Handle(next goyave.HandlerV5) goyave.HandlerV5 {
	return func(response *goyave.ResponseV5, r *goyave.RequestV5) {

		if err := parseQuery(r); err != nil {
			response.Status(http.StatusBadRequest)
			return
		}

		r.Data = nil
		contentType := r.Header().Get("Content-Type")
		if contentType != "" {
			maxSize := int64(m.getMaxUploadSize() * 1024 * 1024)
			maxValueBytes := maxSize
			var bodyBuf bytes.Buffer
			n, err := io.CopyN(&bodyBuf, r.Request().Body, maxValueBytes+1)
			_ = r.Request().Body.Close()
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
					req := r.Request()
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					r.Data, err = generateFlatMap(req, maxSize)
					if err != nil {
						response.Status(http.StatusBadRequest)
					}
				}
			} else {
				response.Status(http.StatusBadRequest)
			}
		}

		if response.GetStatus() != http.StatusBadRequest {
			next(response, r)
		}
	}
}

func (m *Middleware) getMaxUploadSize() float64 {
	if m.MaxUploadSize == 0 {
		return m.Config().GetFloat("server.maxUploadSize")
	}

	return m.MaxUploadSize
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
	request.Form = url.Values{} // Prevent Form from being parse because it would be redundant with our parsing
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

	if request.PostForm != nil {
		flatten(flatMap, request.PostForm)
	}
	if request.MultipartForm != nil {
		flatten(flatMap, request.MultipartForm.Value)

		for field, headers := range request.MultipartForm.File {
			files, err := fsutil.ParseMultipartFiles(headers)
			if err != nil {
				return nil, err
			}
			flatMap[field] = files
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
