package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/samber/lo"
)

// MaxStackDepth the maximum number of frames collected when creating a new Error.
var MaxStackDepth = 50

// Error wraps an errors and attaches callers at the time of creation.
// This implementation provides information for debugging or error reporting.
// It is encouraged to use this type of error everywhere a function can return
// an unexpected error. Functions returning an error that is only indicative should
// not use this error type (e.g.: user errors).
type Error struct {
	reasons      []error
	callers      []uintptr
	callerFrames FrameStack
}

// New create a new `*Error`. Collects the function callers.
//
// If the given reason is already of type `*Error`, returns it without change.
// If the reason is a slice of `error`, joins them using std's `errors.Join`.
//
// If the given reason is `nil`, returns `nil`. If the reason is `[]error` or `[]any`,
// the `nil` elements are ignored.
//
// If the reason is anything other than an `error`, `[]error`, `*Error`, `[]*Error`,
// `[]any`, it will be wrapped in a `Reason` structure, allowing to preserve
// its JSON marshaling behavior.
func New(reason any) error {
	return NewSkip(reason, 3)
}

// NewSkip create a new `*Error`. Collects the function callers, skipping the given
// amount of frames.
//
// If the given reason is already of type `*Error`, returns it without change.
// If the reason is a slice of `error`, joins them using std's `errors.Join`.
//
// If the given reason is `nil`, returns `nil`. If the reason is `[]error` or `[]any`,
// the `nil` elements are ignored.
//
// If the reason is anything other than an `error`, `[]error`, `*Error`, `[]*Error`,
// `[]any`, it will be wrapped in a `Reason` structure, allowing to preserve
// its JSON marshaling behavior.
func NewSkip(reason any, skip int) error {
	if reason == nil {
		return nil
	}
	if r, ok := reason.(*Error); ok {
		return r
	}
	callers := make([]uintptr, 50)
	_ = runtime.Callers(skip, callers)

	return &Error{
		reasons: toErr(reason),
		callers: callers,
	}
}

// Errorf is a shortcut for `errors.New(fmt.Errorf("format", args))`.
// Be careful when using this, this will result in losing the callers of
// the original error if one of the `args` is of type `*errors.Error`.
func Errorf(format string, args ...any) error {
	return NewSkip(fmt.Errorf(format, args...), 3)
}

func toErr(reason any) []error {
	errs := []error{}
	switch r := reason.(type) {
	case error:
		errs = append(errs, r)
	case []error:
		errs = append(errs, lo.Filter(r, func(e error, _ int) bool {
			return e != nil
		})...)
	case []*Error:
		for _, e := range r {
			errs = append(errs, e)
		}
	case []any:
		for _, e := range r {
			if e == nil {
				continue
			}
			errs = append(errs, toErr(e)...)
		}
	default:
		errs = append(errs, Reason{reason: r})
	}
	return errs
}

func (e Error) Error() string {
	if len(e.reasons) == 0 {
		return "goyave.dev/goyave/util/errors.Error: the Error doesn't wrap any reason (empty reasons slice)"
	}
	return strings.Join(lo.Map(e.reasons, func(e error, _ int) string {
		if e == nil {
			return "<nil>"
		}
		return e.Error()
	}), "\n")
}

func (e Error) String() string {
	if len(e.reasons) == 0 {
		return e.Error() + "\n" + e.StackFrames().String()
	}
	if len(e.reasons) == 1 {
		if err, ok := e.reasons[0].(*Error); ok {
			return err.String()
		}
		return e.Error() + "\n" + e.StackFrames().String()
	}
	errs := lo.Map(e.reasons, func(err error, _ int) string {
		if err, ok := err.(*Error); ok {
			return err.String()
		}
		if err == nil {
			return "<nil>\n" + e.StackFrames().String()
		}
		return err.Error() + "\n" + e.StackFrames().String()
	})
	return strings.Join(errs, "\n\n")
}

// FileLine returns the file path and line of the error.
func (e Error) FileLine() string {
	frames := e.StackFrames()
	if len(frames) > 0 {
		f := frames[0]
		return fmt.Sprintf("%s:%d", f.File, f.Line)
	}
	return "[unknown file line]"
}

func (e Error) Unwrap() []error {
	return e.reasons
}

// Len returns the number of underlying reasons.
func (e Error) Len() int {
	return len(e.reasons)
}

// Callers returns the function callers collected at the time of creation of the `Error`.
func (e Error) Callers() []uintptr {
	return e.callers
}

// StackFrames returns the parsed `FrameStack` for this error.
func (e Error) StackFrames() FrameStack {
	if e.callerFrames == nil {
		frames := runtime.CallersFrames(e.callers)
		e.callerFrames = make(FrameStack, 0, len(e.callers))
		for frame, more := frames.Next(); more; frame, more = frames.Next() {
			e.callerFrames = append(e.callerFrames, frame)
		}
	}

	return e.callerFrames
}

// MarshalJSON marshals the error and its underlying reasons.
// The result will be:
// - a string if the reasons slice is empty
// - the marshaled first reason if the reasons slice contains only one error
// - an array of the marshaled reasons otherwise
func (e Error) MarshalJSON() ([]byte, error) {
	if len(e.reasons) == 0 {
		return json.Marshal(e.Error())
	}
	if len(e.reasons) == 1 {
		return e.marshalReason(e.reasons[0])
	}
	marshaledErrors := make([]string, 0, len(e.reasons))
	for _, r := range e.reasons {
		res, err := e.marshalReason(r)
		if err != nil {
			return nil, err
		}
		marshaledErrors = append(marshaledErrors, string(res))
	}
	return []byte(fmt.Sprintf("[%s]", strings.Join(marshaledErrors, ","))), nil
}

func (Error) marshalReason(e error) ([]byte, error) {
	switch err := e.(type) {
	case json.Marshaler, nil:
		return json.Marshal(err)
	default:
		return json.Marshal(err.Error())
	}
}

// FrameStack slice of frames containing information about the stack.
// Can be used to generate a stack trace for debugging, or for error reporting.
type FrameStack []runtime.Frame

func (s FrameStack) String() string {
	return strings.Join(lo.Map(s, func(f runtime.Frame, _ int) string {
		return fmt.Sprintf("%s\n\t%s:%d", f.Function, f.File, f.Line)
	}), "\n")
}

// Reason wrapper around any type of error Reason. This allows json marshaling of the Reason
// instead of losing the original data using `%v` format.
// Calling `Error()` on this structure returns the original data formatted with `%v`.
type Reason struct {
	reason any
}

// Value returns the reason's value.
func (r Reason) Value() any {
	return r.reason
}

func (r Reason) Error() string {
	return fmt.Sprintf("%v", r.reason)
}

// MarshalJSON marshals the wrapped reason.
func (r Reason) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.reason)
}
