package utilities

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func ExtractPullRequestInfo(link string) (owner string, repo string, number int, err error) {
	var u *url.URL
	u, err = url.Parse(link)
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

// Check Error and throw log break programe
func checkErrThrowLog(err error, messages ...string) {
	if len(messages) == 0 {
		messages = []string{"[Redmine] You have a error: %v"}
	}
	if err != nil {
		for _, message := range messages {
			log.Fatalf(message, err)
		}
	}
}

// Parse response body from request
func ParseHttpResponseBody(resp *http.Response) []byte {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErrThrowLog(err)
	return body
}

func ConvertStringToDecimal(char string) int {
	return int([]rune(char)[0])
}

func GetColumnDistance(from, to string) int {
	return ConvertStringToDecimal(to) - ConvertStringToDecimal(from)
}

func GetMinMaxCharacter(chars ...string) (min string, max string) {
	sort.Strings(chars)
	min = chars[0]
	max = chars[len(chars)-1]
	return min, max
}

func GetIDTicket(ticket, splitStr string) int {
	s := strings.Split(ticket, splitStr)

	if len(s) > 1 {
		idTicket, err := strconv.Atoi(strings.Trim(s[1], "/"))
		if err != nil {
			return 0
		}
		return idTicket
	}
	return 0
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
