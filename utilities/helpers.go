package utilities

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func ExtractPullRequestInfo(link string) (owner string, repo string, number int, err error) {
	var u *url.URL
	u, err = url.Parse(link)
	fmt.Println(u.Path)
	if err != nil {
		return
	}
	elms := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(elms) < 4 {
		err = errors.New("Can not parse Github pull request link")
		return
	}
	number, err = strconv.Atoi(elms[3])
	if err != nil {
		return
	}
	owner, repo = elms[0], elms[1]
	return
}

func ArrayValueExists(array []int, neddle int) bool {
	for _, value := range array {
		if value == neddle {
			return true
		}
	}
	return false
}
