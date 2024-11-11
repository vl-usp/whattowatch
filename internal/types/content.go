package types

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
	"time"
)

type ContentItem struct {
	ID           int64
	ContentType  ContentType
	Title        string
	Overview     string
	Popularity   float32
	PosterPath   string
	BackdropPath string
	ReleaseDate  time.Time
	VoteAverage  float32
	VoteCount    int64
	Genres       Genres
	TrailerURL   string
	Counties     []string
}

func SerializeContentItem(c ContentItem) []byte {
	var b bytes.Buffer

	enc := gob.NewEncoder(&b)
	if err := enc.Encode(c); err != nil {
		fmt.Println("Error encoding struct:", err)
		return nil
	}

	return b.Bytes()
}

func UnserializeContentItem(data []byte) (ContentItem, error) {
	b := bytes.NewBuffer(data)

	var ci ContentItem
	dec := gob.NewDecoder(b)
	if err := dec.Decode(&ci); err != nil {
		return ContentItem{}, err
	}

	return ci, nil
}

func (c ContentItem) GetInfo() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("*Название:* %s (%d год", c.Title, c.ReleaseDate.Year()))
	if len(c.Counties) > 0 {
		sb.WriteString(fmt.Sprintf("; %s)\n", strings.Join(c.Counties, ", ")))
	} else {
		sb.WriteString(")\n")
	}
	if len(c.Genres) > 0 {
		sb.WriteString(fmt.Sprintf("*Жанры:* %s\n", c.Genres.String()))
	}
	sb.WriteString(fmt.Sprintf("*Рейтинг:* %s (%d чел.)\n", fmt.Sprintf("%.2f", c.VoteAverage), c.VoteCount))
	if c.Overview != "" {
		sb.WriteString(fmt.Sprintf("*Описание:* %s\n", c.Overview))
	}
	if c.TrailerURL != "" {
		sb.WriteString(fmt.Sprintf("[Ссылка на трейлер](%s)", c.TrailerURL))
	}

	return sb.String()
}

func (c ContentItem) GetShortInfo() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("/%s%d\n", c.ContentType.Sign(), c.ID))
	sb.WriteString(fmt.Sprintf("*Название:* %s (%d год", c.Title, c.ReleaseDate.Year()))
	if len(c.Counties) > 0 {
		sb.WriteString(fmt.Sprintf("; %s)\n", strings.Join(c.Counties, ", ")))
	} else {
		sb.WriteString(")\n")
	}
	sb.WriteString(fmt.Sprintf("*Рейтинг:* %s (%d чел.)\n", fmt.Sprintf("%.2f", c.VoteAverage), c.VoteCount))
	if c.Overview != "" {
		overview := c.Overview
		if len([]rune(overview)) > 500 {
			overview = string([]rune(overview)[:500]) + "..."
		}
		sb.WriteString(fmt.Sprintf("*Описание:* %s\n", overview))
	}

	return sb.String()
}

type Content []ContentItem

func (content Content) IDs() []int64 {
	ids := make([]int64, 0, len(content))
	for _, c := range content {
		ids = append(ids, c.ID)
	}
	return ids
}

func (content Content) RemoveByIDs(ids []int64) Content {
	result := make(Content, 0, len(content))
	filterMap := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		filterMap[id] = struct{}{}
	}

	for _, c := range content {
		if _, ok := filterMap[c.ID]; !ok {
			result = append(result, c)
		}
	}

	return result
}

func (content Content) RemoveDuplicates() Content {
	result := make(Content, 0, len(content))
	ids := make(map[int64]struct{}, len(content))
	for _, c := range content {
		if _, ok := ids[c.ID]; !ok {
			result = append(result, c)
			ids[c.ID] = struct{}{}
		}
	}

	return result
}
