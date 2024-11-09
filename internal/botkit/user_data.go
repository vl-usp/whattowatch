package botkit

import (
	"fmt"

	"github.com/go-telegram/ui/keyboard/reply"
)

type Page int

const (
	MoviePopular Page = iota
	MovieTop
	TVPopular
	TVTop
)

func (p Page) String() string {
	return [...]string{"MoviePopular", "MovieTop", "TVPopular", "TVTop"}[p]
}

func ParsePage(s string) (Page, error) {
	for i, name := range [...]string{"MoviePopular", "MovieTop", "TVPopular", "TVTop"} {
		if name == s {
			return Page(i), nil
		}
	}
	return 0, fmt.Errorf("unknown type")
}

func (p Page) ID() int {
	return int(p)
}

type UserData struct {
	replyKeyboard *reply.ReplyKeyboard

	pagesMap map[Page]int
}

func initUserData(kbFunc keyboardFunc) UserData {
	pagesMap := make(map[Page]int)
	pagesMap[MoviePopular] = 1
	pagesMap[MovieTop] = 1
	pagesMap[TVPopular] = 1
	pagesMap[TVTop] = 1

	return UserData{
		replyKeyboard: kbFunc(),
		pagesMap:      pagesMap,
	}
}
