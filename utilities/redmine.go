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
	redmineUrlIssue = `/issues/`
	redmineUrl      = `/issues.json`
)

type Redmine struct {
	config   *redmineConfig
	url      string
	urlIssue string
}

type redmineConfig struct {
	Token             string `json:"token"`
	Url               string `json:"url"`
	ProjectIdentifier string `json:"project_identifier"`
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
}

type redmineReponse struct {
	Issue redmineIssue `json:"issue"`
}

type redmineArray struct {
	Issues []redmineIssue `json:"issues"`
}

type redmineNotify struct {
	User      string `json:user`
	Status    string `json:status`
	Subject   string `json:subject`
	DoneRatio int    `json:done_ratio`
}

type redmineProjectIssue struct {
	Issues []struct {
		Status struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"status"`
		AssignedTo struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"assigned_to,omitempty"`
	}
}

type redmineProjectMember struct {
	Memberships []struct {
		User struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"user"`
	}
}

type redmineUser struct {
	ID          int       `json:"id"`
	Firstname   string    `json:"firstname"`
	Lastname    string    `json:"lastname"`
	Mail        string    `json:"mail"`
	CreatedOn   time.Time `json:"created_on"`
	LastLoginOn time.Time `json:"last_login_on"`
}

type redmineProject struct {
	Projects []struct {
		ID          int       `json:"id"`
		Name        string    `json:"name"`
		Identifier  string    `json:"identifier"`
		Description string    `json:"description"`
		Status      int       `json:"status"`
		IsPublic    bool      `json:"is_public"`
		CreatedOn   time.Time `json:"created_on"`
		UpdatedOn   time.Time `json:"updated_on"`
	} `json:"projects"`
}

type redmineUserConfig []struct {
	Name      string `json:"name"`
	CwID      string `json:"chatwork_id"`
	RedmineID string `json:"redmine_id"`
}

type redmineNoticeUser struct {
	NameProject        string            `json:"name_project"`
	RedmineUserProject redmineUserConfig `json:"redmine_user_config"`
}

func NewRedmine() *Redmine {
	redmine := &Redmine{}
	redmine.loadConfig()
	redmine.url = redmine.config.Url + redmineUrl
	redmine.urlIssue = redmine.config.Url + redmineUrlIssue
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

func (r *Redmine) GetUserInProgress(date string) map[int]string {
	t := time.Now()
	if date == "" {
		date = t.Format(configs.FORMAT_DATE)
	}

	id := r.GetProjectId()
	resp := SetUpRequestToService("GET", r.config.Url+"issues.json?project_id="+id+"&status_id=open&?created_on=%3E%3D"+date, func(req *http.Request) {
		req.Header.Add("X-Redmine-API-Key", r.config.Token)
	})
	response := redmineProjectIssue{}
	err := json.Unmarshal(ParseHttpResponseBody(resp), &response)
	checkErrThrowLog(err)
	ticketInPr := make(map[int]string)

	for _, elem := range response.Issues {
		if elem.Status.Name == "In Progress" {
			ticketInPr[elem.AssignedTo.ID] = elem.AssignedTo.Name
		}
	}
	return ticketInPr
}

func (r *Redmine) GetUserNotInProgress() {
	resp := SetUpRequestToService("GET", r.config.Url+"projects/"+r.config.ProjectIdentifier+"/memberships.json", func(req *http.Request) {
		req.Header.Add("X-Redmine-API-Key", r.config.Token)
	})
	response := redmineProjectMember{}
	err := json.Unmarshal(ParseHttpResponseBody(resp), &response)
	checkErrThrowLog(err)
	userProject := make(map[int]string)
	userInProgress := r.GetUserInProgress("")

	for _, elem := range response.Memberships {
		if _, ok := userInProgress[elem.User.ID]; !ok {
			userProject[elem.User.ID] = elem.User.Name
		}
	}

	redmineUsers := make(map[string]string)

	for _, elem := range userProject {
		redmineUsers[elem] = elem
	}

	redmineUserConfigs := r.LoadRedmindUserConfig()
	userNotification := redmineUserConfig{}
	for _, elem := range redmineUserConfigs {
		if _, ok := redmineUsers[elem.Name]; ok {
			userNotification = append(userNotification, elem)
		}
	}

	userSendNotice := redmineNoticeUser{}
	userSendNotice.RedmineUserProject = userNotification
	userSendNotice.NameProject = r.GetProjectName()

	chatwork := NewChatwork()
	chatwork.SendNoticeMemberNotHaveTaskInprogeress(userSendNotice)
}

func (r *Redmine) GetProjectId() string {
	project := redmineProject{}
	resp := SetUpRequestToService("GET", r.config.Url+"projects.json", func(req *http.Request) {
		req.Header.Add("X-Redmine-API-Key", r.config.Token)
	})
	err := json.Unmarshal(ParseHttpResponseBody(resp), &project)
	checkErrThrowLog(err)
	var id string
	for _, elem := range project.Projects {
		if elem.Identifier == r.config.ProjectIdentifier {
			id = strconv.Itoa(elem.ID)
		}
	}
	return id
}

func (r *Redmine) GetProjectName() string {
	project := redmineProject{}
	resp := SetUpRequestToService("GET", r.config.Url+"projects.json", func(req *http.Request) {
		req.Header.Add("X-Redmine-API-Key", r.config.Token)
	})
	err := json.Unmarshal(ParseHttpResponseBody(resp), &project)
	checkErrThrowLog(err)
	var name string
	for _, elem := range project.Projects {
		if elem.Identifier == r.config.ProjectIdentifier {
			name = elem.Name
		}
	}
	return name
}

func (r *Redmine) LoadRedmindUserConfig() redmineUserConfig {
	redmineUser := redmineUserConfig{}
	content, e := ioutil.ReadFile(configs.PATH_REDMINE_USER)
	checkErrThrowLog(e)
	err := json.Unmarshal(content, &redmineUser)
	checkErrThrowLog(err)

	return redmineUser
}
