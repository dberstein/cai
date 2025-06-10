package request

import (
	"github.com/dberstein/cai/content"
)

type Request struct {
	Content *content.String
}

func New() *Request {
	return &Request{Content: content.NewString(nil)}
}
