package response

import (
	"fmt"
)

type Response struct {
	Content fmt.Stringer
	err     error
}

func New(content fmt.Stringer, err error) *Response {
	return &Response{
		Content: content,
		err:     err,
	}
}

func (r *Response) SetError(err error) {
	r.err = err
}

func (r *Response) Error() error {
	return r.err
}

func (r *Response) Process() error {
	return nil
}

func (r *Response) String() fmt.Stringer {
	return r.Content
}
