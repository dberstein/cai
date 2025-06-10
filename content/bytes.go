package content

func NewBytes(b *[]byte) *Bytes {
	bs := &Bytes{Content: []byte{}}
	if b != nil {
		bs.Append(b)
	}
	return bs
}

type Bytes struct {
	Content []byte
}

func (b *Bytes) Append(txt *[]byte) {
	b.Content = append(b.Content, *txt...)
}

func (b *Bytes) String() string {
	return string(b.Content)
}

func (b *Bytes) Len() int {
	return len(b.Content)
}

func (b *Bytes) Reset() {
	b.Content = []byte{}
}
