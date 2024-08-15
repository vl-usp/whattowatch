package types

import "fmt"

type ContentType int

const (
	MovieContentType ContentType = iota + 1
	TVContentType
)

func (w ContentType) String() string {
	return [...]string{"MovieContentType", "TVContentType"}[w-1]
}

func ParseContentType(s string) (ContentType, error) {
	for i, name := range [...]string{"MovieContentType", "TVContentType"} {
		if name == s {
			return ContentType(i + 1), nil
		}
	}
	return 0, fmt.Errorf("unknown type")
}

func (w ContentType) EnumIndex() int {
	return int(w)
}

type TMDbContentType struct {
	ID   int
	Name string
}
