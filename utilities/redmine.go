package utilities

import (
	"crypto/tls"
	"encoding/json"
	"github.com/euclid1990/gstats/configs"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"sync"
)

type Redmine struct {
	config *redmineConfig
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
	EstimatedHours float64              `json:"estimated_hours"`
	Parent         redmineRelate        `json:"parent"`
	CustomFields   []redmineCustomField `json:"custom_fields"`
}

type redmineReponse struct {
	Issue redmineIssue `json:"issue"`
}

func NewRedmine() *Redmine {
	redmine := &Redmine{}
	redmine.loadConfig()
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
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, requestErr := http.NewRequest("GET", r.config.Url+strconv.Itoa(id)+".json", nil)
	checkErrThrowLog(requestErr)
	req.Header.Add("X-Redmine-API-Key", r.config.Token)
	req.Header.Add("Content-Type", "application/json")
	resp, responseErr := client.Do(req)
	checkErrThrowLog(responseErr)
	response := redmineReponse{}
	err := json.Unmarshal(ParseHttpResponseBody(resp), &response)
	checkErrThrowLog(err)
	return response
}

func (r *Redmine) GetIds(ids ...int) []redmineReponse {
	count := len(ids)
	if count == 0 {
		return []redmineReponse{}
	}

	count_done := 0
	wg := sync.WaitGroup{}
	idChan := make(chan int)
	arrayRedmine := make([]redmineReponse, count)
	locker := sync.Mutex{}

	go func() {
		for _, id := range ids {
			idChan <- id
		}
	}()
	wg.Add(1)
	for i := 1; i < runtime.GOMAXPROCS(0); i++ {
		go func(idC chan int) {
			for {
				select {
				case id := <-idC:
					locker.Lock()
					arrayRedmine = append(arrayRedmine, r.Get(id))
					locker.Unlock()
					count_done++
					if count_done == count {
						wg.Done()
					}
				}
			}
		}(idChan)
	}

	wg.Wait()

	return arrayRedmine
}
