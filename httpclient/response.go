package httpclient

import (
	"encoding/json"
	"fmt"
)

func (r *Response) JSON(v interface{}) error {
	if err := json.Unmarshal(r.Body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}
	return nil
}

func (r *Response) String() string {
	return string(r.Body)
}

func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

func (r *Response) IsError() bool {
	return r.StatusCode >= 400
}
