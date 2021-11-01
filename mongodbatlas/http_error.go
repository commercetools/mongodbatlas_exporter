package mongodbatlas

/*
* HTTPError provides additional information
* about an HTTP Error Response such as the status code.
* The intention is to use this to propagate status codes
* up to a prometheus collector to report the number of errors.
 */

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
