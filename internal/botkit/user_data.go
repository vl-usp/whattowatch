package botkit

import (
	"whattowatch/internal/types"

	"github.com/go-telegram/ui/keyboard/reply"
)

type Page int

const (
	MoviePopular Page = iota
	MovieTop
	TVPopular
	TVTop
	MovieByGenre
	TVByGenre
)

type UserData struct {
	replyKeyboard *reply.ReplyKeyboard

	pagesMap      map[Page]int
	selectedGenre map[types.ContentType]int
}

func initUserData(kbFunc keyboardFunc) UserData {
	pagesMap := make(map[Page]int)
	pagesMap[MoviePopular] = 1
	pagesMap[MovieTop] = 1
	pagesMap[TVPopular] = 1
	pagesMap[TVTop] = 1
	pagesMap[MovieByGenre] = 1
	pagesMap[TVByGenre] = 1

	selectedGenre := make(map[types.ContentType]int)

	return UserData{
		replyKeyboard: kbFunc(),
		pagesMap:      pagesMap,
		selectedGenre: selectedGenre,
	}
}
