package types

type SourceLinkMap map[int]SourceLink

type SourceLink struct {
	OriginalID int
	SourceID   int
	Page       int
	Title      string
	Url        string
	MovieUrl   string
}
