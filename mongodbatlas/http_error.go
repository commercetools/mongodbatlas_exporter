package mongodbatlas

import "fmt"

type HTTPError struct {
	StatusCode int
	Err        error
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func (e *HTTPError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("%d: %s", e.StatusCode, e.Err.Error())
	}
	return e.Err.Error()
}
