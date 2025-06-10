package content

import (
	"strings"
)

type String struct {
	sb strings.Builder
}

func NewString(s *string) *String {
	ss := &String{sb: strings.Builder{}}
	if s != nil {
		ss.Append(s)
	}
	return ss
}

func (s *String) Append(txt *string) {
	s.sb.WriteString(*txt)
}

func (s *String) String() string {
	return s.sb.String()
}

func (s *String) Len() int {
	return s.sb.Len()
}

func (s *String) Reset() {
	s.sb.Reset()
}
