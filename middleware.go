package goyave

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime/debug"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/helper/filesystem"
	"github.com/System-Glitch/goyave/v2/lang"
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
		defer func() {
			if err := recover(); err != nil {
				ErrLogger.Println(err)
				response.err = err
				if config.GetBool("debug") {
					response.stacktrace = string(debug.Stack())
				}
				response.Status(http.StatusInternalServerError)
			}
		}()

		next(response, r)
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
		maxSize := int64(config.Get("maxUploadSize").(float64) * 1024 * 1024)
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
			if request.httpRequest.Header.Get("Content-Type") == "application/json" {
				request.Data = make(map[string]interface{}, 10)
				if err := json.Unmarshal(bodyBytes, &request.Data); err != nil {
					request.Data = nil
				}
				resetRequestBody(request, bodyBytes)
			} else {
				resetRequestBody(request, bodyBytes)
				request.Data = generateFlatMap(request.httpRequest, maxSize)
				resetRequestBody(request, bodyBytes)
			}
		}

		next(response, request)
	}
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

func generateFlatMap(request *http.Request, maxSize int64) map[string]interface{} {
	var flatMap map[string]interface{} = make(map[string]interface{})
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
		for field, value := range request.Form {
			if len(value) > 1 {
				flatMap[field] = value
			} else {
				flatMap[field] = value[0]
			}
		}
	}
	if request.MultipartForm != nil {
		for field, value := range request.MultipartForm.Value {
			if len(value) > 1 {
				flatMap[field] = value
			} else {
				flatMap[field] = value[0]
			}
		}

		for field := range request.MultipartForm.File {
			flatMap[field] = filesystem.ParseMultipartFiles(request, field)
		}
	}

	// Source form is not needed anymore, clear it.
	request.Form = url.Values{}
	request.PostForm = url.Values{}
	request.MultipartForm = nil

	return flatMap
}

func resetRequestBody(request *Request, bodyBytes []byte) {
	request.httpRequest.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
}

// validateRequestMiddleware is a middleware that validates the request and sends a 422 error code
// if the validation rules are not met.
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
		response.JSON(code, errsBag)
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
			request.Lang = config.GetString("defaultLanguage")
		}
		next(response, request)
	}
}
