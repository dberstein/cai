package request

import (
	"os"

	"github.com/dberstein/cai/content"
)

type Request struct {
	Content *content.String
	Output  *os.File
}

func New() *Request {
	return &Request{Content: content.NewString(nil), Output: os.Stdout}
}
