package types

type ContentStatus struct {
	UserID     int64
	ContentID  int64
	IsViewed   bool
	IsFavorite bool
}
