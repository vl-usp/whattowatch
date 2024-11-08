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

func (w ContentType) ID() int {
	return int(w)
}

func (w ContentType) Sign() string {
	switch w {
	case Movie:
		return "f"
	case TV:
		return "t"
	}
	return ""
}
