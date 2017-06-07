package utilities

import (
	"encoding/json"
	"github.com/euclid1990/gstats/configs"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	REDMINE_URL_ISSUE = `/issues/`
	REDMINE_URL       = `/issues.json`
)

type Redmine struct {
	config   *redmineConfig
	url      string
	urlIssue string
}

type redmineConfig struct {
	Token string `json:"token"`
	Url   string `json:"url"`
}

type redmineRelate struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type redmineCustomField struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type redmineIssue struct {
	Id             int                  `json:"id"`
	Project        redmineRelate        `json:"project"`
	Status         redmineRelate        `json:"status"`
	Author         redmineRelate        `json:"author"`
	AssignTo       redmineRelate        `json:"assigned_to"`
	Description    string               `json:"description"`
	Subject        string               `json:"subject"`
	StartDate      string               `json:"start_date"`
	DueDate        string               `json:"due_date"`
	DoneRatio      int                  `json:"done_ratio"`
	EstimatedHours float64              `json:"estimated_hours"`
	Parent         redmineRelate        `json:"parent"`
	CustomFields   []redmineCustomField `json:"custom_fields"`
	Url            string
}

type redmineReponse struct {
	Issue redmineIssue `json:"issue"`
}

type redmineArray struct {
	Issues []redmineIssue `json:"issues"`
}

type redmineNotify struct {
	User      string `json:"user"`
	Status    string `json:"status"`
	Subject   string `json:"subject"`
	DoneRatio int    `json:"done_ratio"`
}

func NewRedmine() *Redmine {
	redmine := &Redmine{}
	redmine.loadConfig()
	redmine.url = redmine.config.Url + REDMINE_URL
	redmine.urlIssue = redmine.config.Url + REDMINE_URL_ISSUE
	return redmine
}

func (r *Redmine) loadConfig() {
	conf := redmineConfig{}
	content, e := ioutil.ReadFile(configs.PATH_REDMINE_SECRET)
	checkErrThrowLog(e)
	err := json.Unmarshal(content, &conf)
	checkErrThrowLog(err)
	r.config = &conf
}

func (r *Redmine) Get(id int) redmineReponse {
	resp := SetUpRequestToService("GET", r.urlIssue+strconv.Itoa(id)+".json", func(req *http.Request) {
		req.Header.Add("X-Redmine-API-Key", r.config.Token)
	})
	response := redmineReponse{}
	err := json.Unmarshal(ParseHttpResponseBody(resp), &response)
	checkErrThrowLog(err)
	return response
}

func (r *Redmine) GetIssueByIds(ids []int) []redmineReponse {
	count := len(ids)
	if count == 0 {
		return []redmineReponse{}
	}
	wg := sync.WaitGroup{}
	idChan := make(chan int)
	arrayRedmine := make([]redmineReponse, count)
	locker := sync.Mutex{}
	for i := 1; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func(idChan chan int) {
			defer wg.Done()
			for {
				select {
				case id, ok := <-idChan:
					if !ok {
						return
					}
					redmine := r.Get(id)
					locker.Lock()
					arrayRedmine = append(arrayRedmine, redmine)
					locker.Unlock()
				}
			}
		}(idChan)
	}
	go func(ch chan<- int) {
		for _, id := range ids {
			ch <- id
		}
		close(ch)
	}(idChan)
	wg.Wait()
	return arrayRedmine
}

func (r *Redmine) GetIssueByDate(date string) redmineArray {
	t := time.Now()
	if date == "" {
		date = t.Format(configs.FORMAT_DATE)
	}
	resp := SetUpRequestToService("GET", r.url+"?created_on=%3E%3C"+date, func(req *http.Request) {
		req.Header.Add("X-Redmine-API-Key", r.config.Token)
	})
	response := redmineArray{}
	err := json.Unmarshal(ParseHttpResponseBody(resp), &response)
	checkErrThrowLog(err)
	return response
}

func (r *Redmine) NotifyInprogressIssuesToChatwork() []redmineNotify {
	data := r.GetIssueByDate("")
	redmineChan := make(chan redmineIssue)
	arrayNotify := []redmineNotify{}
	locker := sync.Mutex{}
	wg := sync.WaitGroup{}
	for i := 1; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func(redmineChan chan redmineIssue) {
			defer wg.Done()
			for {
				select {
				case issue, ok := <-redmineChan:
					if !ok {
						return
					}
					if issue.Status.Id != configs.ISSUE_STATUS_RESOLVED { /*	issue = 3  is resolved	*/
						temp := redmineNotify{
							User:      issue.AssignTo.Name,
							Status:    issue.Status.Name,
							Subject:   issue.Subject,
							DoneRatio: issue.DoneRatio,
						}
						locker.Lock()
						arrayNotify = append(arrayNotify, temp)
						locker.Unlock()
					}
				}
			}
		}(redmineChan)
	}

	go func(redmineChan chan<- redmineIssue) {
		for _, issue := range data.Issues {
			redmineChan <- issue
		}
		close(redmineChan)
	}(redmineChan)
	wg.Wait()
	chatwork := NewChatwork()
	chatwork.SendInprogressIssuesMessage(arrayNotify)
	return arrayNotify
}

func (r *Redmine) UpdateStoryPoint(loc *Loc) error {
	prs := loc.Pr
	prChan := make(chan PR)
	arrayErr := []error{}
	arrayPR := []PR{}
	locker := sync.Mutex{}
	wg := sync.WaitGroup{}
	for i := 1; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func(prChan chan PR) {
			defer wg.Done()
			for {
				select {
				case pr, ok := <-prChan:
					if !ok {
						return
					}
					issue := r.Get(pr.IDTicket)
					if len(issue.Issue.CustomFields) != 0 {
						point, err := strconv.Atoi(issue.Issue.CustomFields[0].Value)
						pr.Point = point
						if err != nil {
							arrayErr = append(arrayErr, err)
							checkError(err)
						}
					}
					locker.Lock()
					arrayPR = append(arrayPR, pr)
					locker.Unlock()
				}
			}
		}(prChan)
	}

	go func(prChan chan<- PR) {
		for _, pr := range prs {
			prChan <- pr
		}
		close(prChan)
	}(prChan)
	wg.Wait()
	if len(arrayErr) > 0 {
		return arrayErr[1]
	}
	loc.Pr = arrayPR
	return nil
}
