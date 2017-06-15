package utilities

import (
	"encoding/json"
	"fmt"
	"github.com/euclid1990/gstats/configs"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"math"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const (
	DEFAULT_OFFSET = 0
	DEFAULT_LIMIT  = 25
)

type Project struct {
	Identifier  string `json:"project_identifier"`
	Name        string
	Members     *Members
	redmine     *Redmine
	Issues      []redmineIssue `json:"issues"`
	TotalIssues int            `json:"total_count"`
}

func NewProject(redmine *Redmine) (*Project, error) {
	project := &Project{}
	raw, err := ioutil.ReadFile(configs.PATH_PROJECT)
	if err != nil {
		return project, err
	}
	err = json.Unmarshal(raw, &project)
	if err != nil {
		return project, err
	}
	project.redmine = redmine
	return project, nil
}

func (project *Project) GetOverdueIssues(offset, limit int) (Project, error) {
	t := time.Now().AddDate(0, 0, -1)
	date := t.Format(configs.FORMAT_DATE)
	identifier := project.Identifier
	url := fmt.Sprintf("%s?project_id=%s&status_id=%s,%s&due_date=%s&offset=%s&limit=%s", project.redmine.url, identifier, strconv.Itoa(configs.ISSUE_STATUS_NEW), strconv.Itoa(configs.ISSUE_STATUS_INPROGRESS), "%3C%3D"+date, strconv.Itoa(offset), strconv.Itoa(limit))

	resp := SetUpRequestToService("GET", url, func(req *http.Request) {
		req.Header.Add("X-Redmine-API-Key", project.redmine.config.Token)
	})
	response := Project{}
	err := json.Unmarshal(ParseHttpResponseBody(resp), &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (project *Project) appendIssuesToMember(overdueIssues map[int][]redmineIssue, issues []redmineIssue) map[int][]redmineIssue {
	for _, issue := range issues {
		issue.Url = fmt.Sprintf("%s%s", project.redmine.urlIssue, strconv.Itoa(issue.Id))
		overdueIssues[issue.AssignTo.Id] = append(overdueIssues[issue.AssignTo.Id], issue)
		// Get Project Name
		project.Name = issue.Project.Name
	}
	return overdueIssues
}

func (project *Project) GetTotalOverdueIssues() int {
	overdueIssues, err := project.GetOverdueIssues(DEFAULT_OFFSET, DEFAULT_LIMIT)
	if err != nil {
		return 0
	}
	return overdueIssues.TotalIssues
}

func (project *Project) GetAllOverdueIssues() (map[int][]redmineIssue, error) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	totalIssues := project.GetTotalOverdueIssues()
	limit := DEFAULT_LIMIT
	pages := int(math.Floor(float64(totalIssues / limit)))
	overdueIssuesByMembers := make(map[int][]redmineIssue)

	if pages > 0 {
		eg := errgroup.Group{}
		for i := 0; i < pages; i++ {
			page := i
			offset := page * limit
			eg.Go(func() error {
				issues, err := project.GetOverdueIssues(offset, limit)
				if err != nil {
					return err
				}
				overdueIssuesByMembers = project.appendIssuesToMember(overdueIssuesByMembers, issues.Issues)
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			return nil, err
		}
	}
	return overdueIssuesByMembers, nil
}

func (project *Project) NotifyOverdueIssuesToChatwork() error {
	overdueIssues, err := project.GetAllOverdueIssues()
	if err != nil {
		return err
	}
	members, err := NewMembers()
	if err != nil {
		return err
	}

	var listMembers []Member
	for redmineID, issues := range overdueIssues {
		newMem := Member{
			RedmineID: redmineID,
			Issues:    issues,
		}
		members.updateMemberByRedmineId(&newMem)
		if newMem.ChatworkID > 0 {
			listMembers = append(listMembers, newMem)
		}
	}
	if len(listMembers) > 0 {
		members.List = listMembers
		project.Members = members
	}
	chatwork := NewChatwork()
	err = chatwork.SendOverdueIssuesMessage(project)
	if err != nil {
		return err
	}
	return nil
}
