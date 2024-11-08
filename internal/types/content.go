package types

import (
	"database/sql"
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

func SerializeContentItemKey(c ContentItem) string {
	return fmt.Sprintf("%d_%d", c.ID, c.ContentType.ID())
}

func UnserializeContentItemKey(str string) (ContentItem, error) {
	var id int64
	var ct int
	_, err := fmt.Sscanf(str, "%d_%d", &id, &ct)
	if err != nil {
		return ContentItem{}, err
	}
	return ContentItem{ID: id, ContentType: ContentType(ct)}, nil
}

func (c ContentItem) String() string {
	overview := c.Overview
	if overview == "" {
		overview = "Описание отсутствует"
	}

	return fmt.Sprintf(
		"*ID:* /%s%d\n\n*Название:* %s\n\n*Жанры:* %s\n\n*Дата выхода:* %s\n\n*Популярность:* %s\n\n*Рейтинг:* %s\n\n*Количество оценок:* %d\n\n*Описание:* %s",
		c.ContentType.Sign(),
		c.ID,
		utils.EscapeString(c.Title),
		c.Genres.String(),
		utils.EscapeString(c.ReleaseDate.Time.Format("02.01.2006")),
		utils.EscapeString(fmt.Sprintf("%.2f", c.Popularity)),
		utils.EscapeString(fmt.Sprintf("%.2f", c.VoteAverage)),
		c.VoteCount,
		utils.EscapeString(overview),
	)
}

func (c ContentItem) ShortString() string {
	overview := c.Overview
	if overview == "" {
		overview = "Описание отсутствует"
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

func (content Content) Print(title string) string {
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
