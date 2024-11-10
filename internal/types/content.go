package types

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"strings"
	"whattowatch/internal/utils"
)

type ContentItem struct {
	ID          int64
	ContentType ContentType
	Title       string
	Overview    string
	Popularity  float32
	PosterPath  string
	ReleaseDate sql.NullTime
	VoteAverage float32
	VoteCount   int64
	Genres      Genres
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
	overview := c.Overview
	if overview == "" {
		overview = "Описание отсутствует"
	}
	// if len([]rune(overview)) > 500 {
	// 	overview = string([]rune(overview)[:500]) + "..."
	// }

	return fmt.Sprintf(
		"*ID:* /%s%d\n\n*Название:* %s\n\n*Жанры:* %s\n\n*Дата выхода:* %s\n\n*Популярность:* %s\n\n*Рейтинг:* %s\n\n*Количество оценок:* %d\n\n*Описание:* %s",
		c.ContentType.Sign(),
		c.ID,
		c.Title,
		c.Genres.String(),
		c.ReleaseDate.Time.Format("02.01.2006"),
		fmt.Sprintf("%.2f", c.Popularity),
		fmt.Sprintf("%.2f", c.VoteAverage),
		c.VoteCount,
		overview,
	)
}

func (c ContentItem) GetShortInfo() string {
	overview := c.Overview
	if overview == "" {
		overview = "Описание отсутствует"
	}
	if len([]rune(overview)) > 500 {
		overview = string([]rune(overview)[:500]) + "..."
	}

	return fmt.Sprintf(
		"*ID:* /%s%d\n\n*Название:* %s\n\n*Дата выхода:* %s\n\n*Рейтинг:* %s\n\n*Описание:* %s",
		c.ContentType.Sign(),
		c.ID,
		utils.EscapeString(c.Title),
		utils.EscapeString(c.ReleaseDate.Time.Format("02.01.2006")),
		utils.EscapeString(fmt.Sprintf("%.2f", c.VoteAverage)),
		utils.EscapeString(overview),
	)
}

type Content []ContentItem

func (content Content) GetInfo(title string) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", title))
	for _, c := range content {
		switch c.ContentType {
		case Movie:
			builder.WriteString("/f")
		case TV:
			builder.WriteString("/t")
		}
		builder.WriteString(fmt.Sprintf("%d %s (год: %d; популярность: %f; оценка: %f)\n", c.ID, c.Title, c.ReleaseDate.Time.Year(), c.Popularity, c.VoteAverage))
	}
	return builder.String()
}

type IDsWithGenreIDs struct {
	ID   int64
	GIDs []int64
}

func (content Content) IDs() []int64 {
	ids := make([]int64, 0, len(content))
	for _, c := range content {
		ids = append(ids, c.ID)
	}
	return ids
}

func (content Content) IDsWithGenres() []IDsWithGenreIDs {
	result := make([]IDsWithGenreIDs, 0, len(content))

	for _, c := range content {
		result = append(result, IDsWithGenreIDs{
			ID:   c.ID,
			GIDs: c.Genres.GetIDs(),
		})
	}

	return result
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
