package utilities

import (
	"encoding/json"
	"github.com/euclid1990/gstats/configs"
	"io/ioutil"
)

type Member struct {
	ChatworkID int    `json:"chatwork_id"`
	RedmineID  int    `json:"redmine_id"`
	Name       string `json:"name"`
	Role       int    `json:"project_role"`
	Issues     []redmineIssue
}

type Members struct {
	LeaderID   int
	LeaderName string
	List       []Member
}

func NewMembers() (*Members, error) {
	members := &Members{}
	raw, err := ioutil.ReadFile(configs.PATH_MEMBERS)
	if err != nil {
		return members, err
	}

	var data []Member
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return members, err
	}
	members.List = data
	return members, nil
}

func (members *Members) updateMemberByRedmineId(mem *Member) {
	var leaderID int
	var leaderName string
	for _, member := range members.List {
		if mem.RedmineID == member.RedmineID {
			// Update Member Info
			mem.ChatworkID = member.ChatworkID
			mem.Role = member.Role
			mem.Name = member.Name
		}
		// Get Team Leader Chatwork ID
		if member.Role == configs.MEMBER_PROJECT_ROLE_LEADER {
			leaderID = member.ChatworkID
			leaderName = member.Name
		}
	}
	members.LeaderID = leaderID
	members.LeaderName = leaderName
}
