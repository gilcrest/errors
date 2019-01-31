package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/rs/zerolog/log"
)

// hError represents an HTTP handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type hError interface {
	error
	Status() int
	ErrKind() string
	ErrParam() string
	ErrCode() string
}

// HTTPErr represents an error with an associated HTTP status code.
type HTTPErr struct {
	HTTPStatusCode int
	Kind           Kind
	Param          Parameter
	Code           Code
	Err            error
}

// Allows HTTPErr to satisfy the error interface.
func (hse HTTPErr) Error() string {
	return hse.Err.Error()
}

// SetErr creates an error type and adds it to the struct
func (hse *HTTPErr) SetErr(s string) {
	hse.Err = Str(s)
}

// ErrKind returns a string denoting the "kind" of error
func (hse HTTPErr) ErrKind() string {
	return hse.Kind.String()
}

// ErrParam returns a string denoting the "kind" of error
func (hse HTTPErr) ErrParam() string {
	return string(hse.Param)
}

// ErrCode returns a string denoting the "kind" of error
func (hse HTTPErr) ErrCode() string {
	return string(hse.Code)
}

// Status Returns an HTTP Status Code.
func (hse HTTPErr) Status() int {
	return hse.HTTPStatusCode
}

type errResponse struct {
	Error svcError `json:"error"`
}

type svcError struct {
	Kind    string `json:"kind,omitempty"`
	Code    string `json:"code,omitempty"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message,omitempty"`
}

// HTTPError takes a writer and an error, performs a type switch to
// determine if the type is an HTTPError (which meets the Error interface
// as defined in this package), then sends the Error as a response to the
// client. If the type does not meet the Error interface as defined in this
// package, then a proper error is still formed and sent to the client,
// however, the Kind and Code will be Unanticipated.
func HTTPError(w http.ResponseWriter, err error) {
	const op Op = "errors.httpError"

	if err != nil {
		// We perform a "type switch" https://tour.golang.org/methods/16
		// to determine the interface value type
		switch e := err.(type) {
		// If the interface value is of type Error (not a typical error, but
		// the Error interface defined above), then
		case hError:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)

			er := errResponse{
				Error: svcError{
					Kind:    e.ErrKind(),
					Code:    e.ErrCode(),
					Param:   e.ErrParam(),
					Message: e.Error(),
				},
			}

			// Marshal errResponse struct to JSON for the response body
			errJSON, _ := json.MarshalIndent(er, "", "    ")

			sendError(w, string(errJSON), e.Status())

		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			cd := http.StatusInternalServerError
			er := errResponse{
				Error: svcError{
					Kind:    Unanticipated.String(),
					Code:    "Unanticipated",
					Message: "Unexpected error - contact support",
				},
			}

			log.Error().Msgf("Unknown Error - HTTP %d - %s", cd, err.Error())

			// Marshal errResponse struct to JSON for the response body
			errJSON, _ := json.MarshalIndent(er, "", "    ")

			sendError(w, string(errJSON), cd)
		}
	}
}

// Taken from standard library, but changed to send application/json as header
// Error replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
// The error message should be json.
func sendError(w http.ResponseWriter, error string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, error)
}

// RE builds an HTTP Response error value from its arguments.
// There must be at least one argument or RE panics.
// The type of each argument determines its meaning.
// If more than one argument of a given type is presented,
// only the last one is recorded.
//
// The types are:
func RE(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to errors.RE with no arguments")
	}
	e := &HTTPErr{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case int:
			e.HTTPStatusCode = arg
		case Kind:
			e.Kind = arg
		case string:
			e.Code = Code(arg)
		case Code:
			e.Code = arg
		case Parameter:
			e.Param = arg
		case *Error:
			// Make a copy
			copy := *arg
			e.Err = &copy
		case error:
			e.Err = arg
		default:
			_, file, line, _ := runtime.Caller(1)
			log.Error().Msgf("errors.E: bad call from %s:%d: %v", file, line, args)
			return Errorf("unknown type %T, value %v in error call", arg, arg)
		}
	}

	return e
}
