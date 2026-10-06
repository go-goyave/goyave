package goyave

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testError struct{}

func (testError) Error() string {
	return "test error"
}

func TestClientError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc         string
		err          error
		wantError    string
		wantMessage  string
		wantCode     int
		wantSentinel error
	}{
		{
			desc:         "Conflict_empty_message",
			err:          Conflict(""),
			wantError:    "client error 409: " + http.StatusText(http.StatusConflict),
			wantMessage:  "",
			wantCode:     http.StatusConflict,
			wantSentinel: ErrConflict,
		},
		{
			desc:         "Conflict",
			err:          Conflict("custom message"),
			wantError:    "client error 409: custom message",
			wantMessage:  "custom message",
			wantCode:     http.StatusConflict,
			wantSentinel: ErrConflict,
		},
		{
			desc:         "NotAcceptable_empty_message",
			err:          NotAcceptable(""),
			wantError:    "client error 406: " + http.StatusText(http.StatusNotAcceptable),
			wantMessage:  "",
			wantCode:     http.StatusNotAcceptable,
			wantSentinel: ErrNotAcceptable,
		},
		{
			desc:         "NotAcceptable",
			err:          NotAcceptable("custom message"),
			wantError:    "client error 406: custom message",
			wantMessage:  "custom message",
			wantCode:     http.StatusNotAcceptable,
			wantSentinel: ErrNotAcceptable,
		},
		{
			desc:         "NotFound_empty_message",
			err:          NotFound(""),
			wantError:    "client error 404: " + http.StatusText(http.StatusNotFound),
			wantMessage:  "",
			wantCode:     http.StatusNotFound,
			wantSentinel: ErrNotFound,
		},
		{
			desc:         "NotFound",
			err:          NotFound("custom message"),
			wantError:    "client error 404: custom message",
			wantMessage:  "custom message",
			wantCode:     http.StatusNotFound,
			wantSentinel: ErrNotFound,
		},
		{
			desc:         "UnprocessableEntity_empty_message",
			err:          UnprocessableEntity(""),
			wantError:    "client error 422: " + http.StatusText(http.StatusUnprocessableEntity),
			wantMessage:  "",
			wantCode:     http.StatusUnprocessableEntity,
			wantSentinel: ErrUnprocessableEntity,
		},
		{
			desc:         "UnprocessableEntity",
			err:          UnprocessableEntity("custom message"),
			wantError:    "client error 422: custom message",
			wantMessage:  "custom message",
			wantCode:     http.StatusUnprocessableEntity,
			wantSentinel: ErrUnprocessableEntity,
		},
		{
			desc:         "Locked_empty_message",
			err:          Locked(""),
			wantError:    "client error 423: " + http.StatusText(http.StatusLocked),
			wantMessage:  "",
			wantCode:     http.StatusLocked,
			wantSentinel: ErrLocked,
		},
		{
			desc:         "Locked",
			err:          Locked("custom message"),
			wantError:    "client error 423: custom message",
			wantMessage:  "custom message",
			wantCode:     http.StatusLocked,
			wantSentinel: ErrLocked,
		},
		{
			desc:         "Forbidden_empty_message",
			err:          Forbidden(""),
			wantError:    "client error 403: " + http.StatusText(http.StatusForbidden),
			wantMessage:  "",
			wantCode:     http.StatusForbidden,
			wantSentinel: ErrForbidden,
		},
		{
			desc:         "Forbidden",
			err:          Forbidden("custom message"),
			wantError:    "client error 403: custom message",
			wantMessage:  "custom message",
			wantCode:     http.StatusForbidden,
			wantSentinel: ErrForbidden,
		},
		{
			desc:         "Unauthorized_empty_message",
			err:          Unauthorized(""),
			wantError:    "client error 401: " + http.StatusText(http.StatusUnauthorized),
			wantMessage:  "",
			wantCode:     http.StatusUnauthorized,
			wantSentinel: ErrUnauthorized,
		},
		{
			desc:         "Unauthorized",
			err:          Unauthorized("custom message"),
			wantError:    "client error 401: custom message",
			wantMessage:  "custom message",
			wantCode:     http.StatusUnauthorized,
			wantSentinel: ErrUnauthorized,
		},
		{
			desc:         "BadRequest_empty_message",
			err:          BadRequest(""),
			wantError:    "client error 400: " + http.StatusText(http.StatusBadRequest),
			wantMessage:  "",
			wantCode:     http.StatusBadRequest,
			wantSentinel: ErrBadRequest,
		},
		{
			desc:         "BadRequest",
			err:          BadRequest("custom message"),
			wantError:    "client error 400: custom message",
			wantMessage:  "custom message",
			wantCode:     http.StatusBadRequest,
			wantSentinel: ErrBadRequest,
		},
		{
			desc:        "Generic_empty_message",
			err:         NewClientError(http.StatusMethodNotAllowed, ""),
			wantError:   "client error 405: " + http.StatusText(http.StatusMethodNotAllowed),
			wantMessage: "",
			wantCode:    http.StatusMethodNotAllowed,
		},
		{
			desc:        "Generic",
			err:         NewClientError(http.StatusMethodNotAllowed, "custom message"),
			wantError:   "client error 405: custom message",
			wantMessage: "custom message",
			wantCode:    http.StatusMethodNotAllowed,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			require.Implements(t, (*ClientError)(nil), c.err)
			clientErr, ok := errors.AsType[ClientError](c.err)
			require.True(t, ok)
			assert.Equal(t, c.wantError, clientErr.Error())
			assert.Equal(t, c.wantMessage, clientErr.Message())
			assert.Equal(t, c.wantCode, clientErr.Code())
			if c.wantSentinel != nil {
				assert.ErrorIs(t, c.err, c.wantSentinel)
			}
			// errors.Is with a generic client error should also work
			genericClientErr := NewClientError(clientErr.Code(), clientErr.Message())
			assert.ErrorIs(t, c.err, genericClientErr)
			// Changing the code should drop the compatibility between the two errors
			genericClientErr.(*clientError).code = 0
			assert.NotErrorIs(t, c.err, genericClientErr)
			assert.NotErrorIs(t, c.err, errors.New(clientErr.Message())) // For the sake of coverage...
		})
	}

	t.Run("Generic_is_as", func(t *testing.T) {
		cases := []struct {
			desc        string
			code        int
			message     string
			wantErrType error
			asFn        func(error) (ClientError, bool)
			want        bool
		}{
			{
				desc:        "same_ptr",
				code:        http.StatusMethodNotAllowed,
				message:     "custom message",
				wantErrType: &clientError{code: http.StatusMethodNotAllowed},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*clientError](err)
				},
				want: true,
			},
			{
				desc:        "same_no_ptr",
				code:        http.StatusMethodNotAllowed,
				message:     "custom message",
				wantErrType: clientError{code: http.StatusMethodNotAllowed},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*clientError](err)
				},
				want: true,
			},
			{
				desc:        "as_interface",
				code:        http.StatusMethodNotAllowed,
				message:     "custom message",
				wantErrType: &clientError{code: http.StatusMethodNotAllowed},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[ClientError](err)
				},
				want: true,
			},
			{
				desc:        "Conflict",
				code:        http.StatusConflict,
				message:     "custom message",
				wantErrType: &ConflictError{},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*ConflictError](err)
				},
				want: true,
			},
			{
				desc:        "NotAcceptable",
				code:        http.StatusNotAcceptable,
				message:     "custom message",
				wantErrType: &NotAcceptableError{},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*NotAcceptableError](err)
				},
				want: true,
			},
			{
				desc:        "NotFound",
				code:        http.StatusNotFound,
				message:     "custom message",
				wantErrType: &NotFoundError{},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*NotFoundError](err)
				},
				want: true,
			},
			{
				desc:        "UnprocessableEntity",
				code:        http.StatusUnprocessableEntity,
				message:     "custom message",
				wantErrType: &UnprocessableEntityError{},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*UnprocessableEntityError](err)
				},
				want: true,
			},
			{
				desc:        "Locked",
				code:        http.StatusLocked,
				message:     "custom message",
				wantErrType: &LockedError{},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*LockedError](err)
				},
				want: true,
			},
			{
				desc:        "Forbidden",
				code:        http.StatusForbidden,
				message:     "custom message",
				wantErrType: &ForbiddenError{},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*ForbiddenError](err)
				},
				want: true,
			},
			{
				desc:        "Unauthorized",
				code:        http.StatusUnauthorized,
				message:     "custom message",
				wantErrType: &UnauthorizedError{},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*UnauthorizedError](err)
				},
				want: true,
			},
			{
				desc:        "BadRequest",
				code:        http.StatusBadRequest,
				message:     "custom message",
				wantErrType: &BadRequestError{},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*BadRequestError](err)
				},
				want: true,
			},
			{
				desc:        "as_type_not_matching_code",
				code:        http.StatusMethodNotAllowed,
				message:     "custom message",
				wantErrType: &clientError{code: http.StatusMethodNotAllowed},
				asFn: func(err error) (ClientError, bool) {
					return errors.AsType[*BadRequestError](err)
				},
				want: false,
			},
			{
				desc:        "is_not_matching_code",
				code:        http.StatusMethodNotAllowed,
				message:     "custom message",
				wantErrType: &clientError{code: http.StatusMethodNotAllowed},
				asFn: func(err error) (ClientError, bool) {
					ok := errors.Is(err, &clientError{code: http.StatusForbidden})
					return nil, ok
				},
				want: false,
			},
			{
				desc:        "unrelated_type",
				code:        http.StatusMethodNotAllowed,
				message:     "custom message",
				wantErrType: &clientError{code: http.StatusMethodNotAllowed},
				asFn: func(err error) (ClientError, bool) {
					_, ok := errors.AsType[*testError](err)
					return nil, ok
				},
				want: false,
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				t.Parallel()
				clientErr := NewClientError(c.code, c.message)
				assert.ErrorIs(t, clientErr, c.wantErrType)

				result, ok := c.asFn(clientErr)
				require.Equal(t, c.want, ok)
				if ok {
					assert.Equal(t, c.message, result.Message())
					assert.Equal(t, c.code, result.Code())
				}
			})
		}
	})
}
