package utils

import (
	"database/sql"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func EscapeString(s string) string {
	m := map[string]string{
		"_": "\\_",
		// "*":      "\\*",
		"[":      "\\[",
		"]":      "\\]",
		"(":      "\\(",
		")":      "\\)",
		"~":      "\\~",
		"`":      "\\`",
		">":      "\\>",
		"#":      "\\#",
		"+":      "\\+",
		"-":      "\\-",
		"=":      "\\=",
		"|":      "\\|",
		"{":      "\\{",
		"}":      "\\}",
		".":      "\\.",
		"!":      "\\!",
		"\u00a0": " ",
		":":      "\\:",
		"–":      "\\-",
		"«":      "\"",
		"»":      "\"",
		",":      "\\,",
		"?":      "\\?",
	}
	for k, v := range m {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}

func ParseCommand(s string) (string, []string, error) {
	if s[0] != '/' {
		return "", nil, fmt.Errorf("command %s should be started with '/': %s", s, s)
	}
	arr := strings.SplitN(s, " ", 2)
	if len(arr) < 2 {
		return arr[0], nil, fmt.Errorf("command %s should have arguments splitted by comma", s)
	}
	return arr[0], strings.Split(arr[1], ", "), nil
}

func MapToSlice[K comparable, V any](m map[K]V) []V {
	s := make([]V, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}

func GetReleaseDate(in string) (sql.NullTime, error) {
	relesaseDate, err := time.Parse("2006-01-02", in)
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("parse release date from %s error: %s", in, err.Error())
	}
	return sql.NullTime{Time: relesaseDate, Valid: true}, nil
}

func HandlePage(page int, order string) int {
	if order == "prev" {
		page--
	} else if order == "next" {
		page++
	}

	if page <= 0 {
		return 1
	}

	return page
}

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func GetMyIP() string {
	cmd := exec.Command("wget", "-qO-", "eth0.me")
	stdout, err := cmd.Output()

	if err != nil {
		return err.Error()
	}

	return string(stdout)
}

func PingHost(host string, port int) error {
	timeout := time.Duration(1 * time.Second)
	conn, err := net.DialTimeout("tcp", host+":"+strconv.Itoa(port), timeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func IntSliceToStringSlice(i []int) []string {
	s := make([]string, 0, len(i))
	for _, v := range i {
		s = append(s, strconv.Itoa(v))
	}
	return s
}
