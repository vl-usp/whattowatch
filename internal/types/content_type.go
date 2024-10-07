package types

import "fmt"

type ContentType int

const (
	Movie ContentType = iota + 1
	TV
)

func (w ContentType) String() string {
	return [...]string{"Movie", "TV"}[w-1]
}

func ParseContentType(s string) (ContentType, error) {
	for i, name := range [...]string{"Movie", "TV"} {
		if name == s {
			return ContentType(i + 1), nil
		}
	}
	return 0, fmt.Errorf("unknown type")
}

func (w ContentType) EnumIndex() int {
	return int(w)
}
