package goyave

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"

	"goyave.dev/goyave/v4/lang"
	"goyave.dev/goyave/v4/util/fsutil"
)

// Middleware function generating middleware handler function.
//
// Request data is available to middleware, but bear in mind that
// it had not been validated yet. That means that you can modify or
// filter data. (Trim strings for example)
type Middleware func(Handler) Handler

// recoveryMiddleware is a middleware that recovers from panic and sends a 500 error code.
// If debugging is enabled in the config and the default status handler for the 500 status code
// had not been changed, the error is also written in the response.
func recoveryMiddleware(next Handler) Handler {
	return func(response *Response, r *Request) {
		panicked := true
		defer func() {
			if err := recover(); err != nil || panicked {
				ErrLogger.Println(err)
				response.err = err
				response.stacktrace = string(debug.Stack())
				response.Status(http.StatusInternalServerError)
			}
		}()

		next(response, r)
		panicked = false
	}
}

// parseRequestMiddleware is a middleware that parses the request data.
//
// If the parsing fails, the request's data is set to nil. If it succeeds
// and there is no data, the request's data is set to an empty map.
//
// If the "Content-Type: application/json" header is set, the middleware
// will attempt to unmarshal the request's body.
//
// This middleware doesn't drain the request body to maximize compatibility
// with native handlers.
//
// The maximum length of the data is limited by the "maxUploadSize" config entry.
// If a request exceeds the maximum size, the middleware doesn't call "next()" and
// sets the response status code to "413 Payload Too Large".
func parseRequestMiddleware(next Handler) Handler {
	return func(response *Response, request *Request) {

		request.Data = nil
		contentType := request.httpRequest.Header.Get("Content-Type")
		if contentType == "" {
			// If the Content-Type is not set, don't parse body
			request.httpRequest.Body.Close()
			request.Data = make(map[string]interface{})
			if err := parseQuery(request); err != nil {
				request.Data = nil
			}
		} else {
			maxSize := maxPayloadSize
			maxValueBytes := maxSize
			var bodyBuf bytes.Buffer
			n, err := io.CopyN(&bodyBuf, request.httpRequest.Body, maxValueBytes+1)
			request.httpRequest.Body.Close()
			if err == nil || err == io.EOF {
				maxValueBytes -= n
				if maxValueBytes < 0 {
					response.Status(http.StatusRequestEntityTooLarge)
					return
				}

				bodyBytes := bodyBuf.Bytes()
				if strings.HasPrefix(contentType, "application/json") {
					request.Data = make(map[string]interface{}, 10)
					if err := parseQuery(request); err != nil {
						request.Data = nil
					} else {
						if err := json.Unmarshal(bodyBytes, &request.Data); err != nil {
							request.Data = nil
						}
					}
					resetRequestBody(request, bodyBytes)
				} else {
					resetRequestBody(request, bodyBytes)
					request.Data = generateFlatMap(request.httpRequest, maxSize)
					resetRequestBody(request, bodyBytes)
				}
			}
		}

		next(response, request)
	}
}

func generateFlatMap(request *http.Request, maxSize int64) map[string]interface{} {
	flatMap := make(map[string]interface{})
	err := request.ParseMultipartForm(maxSize)

	if err != nil {
		if err == http.ErrNotMultipart {
			if err := request.ParseForm(); err != nil {
				return nil
			}
		} else {
			return nil
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

	return flatMap
}

func flatten(dst map[string]interface{}, values url.Values) {
	for field, value := range values {
		if len(value) > 1 {
			dst[field] = value
		} else {
			dst[field] = value[0]
		}
	}
}

func parseQuery(request *Request) error {
	queryParams, err := url.ParseQuery(request.URI().RawQuery)
	if err == nil {
		flatten(request.Data, queryParams)
	}
	return err
}

func resetRequestBody(request *Request, bodyBytes []byte) {
	request.httpRequest.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
}

// corsMiddleware is the middleware handling CORS, using the options set in the router.
// This middleware is automatically inserted first to the router's list of middleware
// if the latter has defined CORS Options.
func corsMiddleware(next Handler) Handler {
	return func(response *Response, request *Request) {
		if request.corsOptions == nil {
			next(response, request)
			return
		}

		options := request.corsOptions
		headers := response.Header()
		requestHeaders := request.Header()

		options.ConfigureCommon(headers, requestHeaders)

		if request.Method() == http.MethodOptions && requestHeaders.Get("Access-Control-Request-Method") != "" {
			options.HandlePreflight(headers, requestHeaders)
			if options.OptionsPassthrough {
				next(response, request)
			} else {
				response.WriteHeader(http.StatusNoContent)
			}
		} else {
			next(response, request)
		}
	}
}

// validateRequestMiddleware is a middleware that validates the request.
// If validation is not rules are not met, sets the response status to 422 Unprocessable Entity
// or 400 Bad Request and the response error (which can be retrieved with `GetError()`) to the
// `validation.Errors` returned by the validator.
// This data can then be used in a status handler.
func validateRequestMiddleware(next Handler) Handler {
	return func(response *Response, r *Request) {
		errsBag := r.validate()
		if errsBag == nil {
			next(response, r)
			return
		}

		var code int
		if r.Data == nil {
			code = http.StatusBadRequest
		} else {
			code = http.StatusUnprocessableEntity
		}
		response.err = errsBag
		response.Status(code)
	}
}

// languageMiddleware is a middleware that sets the language of a request.
//
// Uses the "Accept-Language" header to determine which language to use. If
// the header is not set or the language is not available, uses the default
// language as fallback.
//
// If "*" is provided, the default language will be used.
// If multiple languages are given, the first available language will be used,
// and if none are available, the default language will be used.
// If no variant is given (for example "en"), the first available variant will be used.
// For example, if "en-US" and "en-UK" are available and the request accepts "en",
// "en-US" will be used.
func languageMiddleware(next Handler) Handler {
	return func(response *Response, request *Request) {
		if header := request.Header().Get("Accept-Language"); len(header) > 0 {
			request.Lang = lang.DetectLanguage(header)
		} else {
			request.Lang = defaultLanguage
		}
		next(response, request)
	}
}
