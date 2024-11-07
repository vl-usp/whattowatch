package types

type ContentStatus struct {
	UserID      int64
	ContentID   int64
	ContentType ContentType
	IsViewed    bool
	IsFavorite  bool
}
