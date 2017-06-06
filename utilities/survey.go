package utilities

import (
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey"
	"github.com/euclid1990/gstats/configs"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type SetupGithubSecret struct {
	ClientId     string `survey:"githubClientId"`
	ClientSecret string `survey:"githubClientSecret"`
}

type SetupGoogleSecret struct {
	ClientId     string `survey:"googleClientId"`
	ClientSecret string `survey:"googleClientSecret"`
}

type SetupChatworkSecret struct {
	Token  string `survey:"chatworkToken"`
	RoomId string `survey:"chatworkRoomId"`
}

type SetupSpreadSheetsSecret struct {
	ID        string `survey:"id"`
	SheetLoc  string `survey:"sheetLoc"`
	CTicket   string `survey:"cTicket"`
	CGithub   string `survey:"cGithub"`
	CPoint    string `survey:"cPoint"`
	CLoc      string `survey:"cLoc"`
	CRowStart string `survey:"cRowStart"`
}

type SetupRedmineSecret struct {
	Token             string `survey:"redmineToken"`
	Url               string `survey:"redmineUrl"`
	ProjectIdentifier string `survey:"redmineProjectIdentifier"`
}

type SetupNumberSpread struct {
	Number string `survey:"numberSpreadQs"`
}

type SetupMember struct {
	ChatworkId  string `survey:"chatworkId"`
	RedmineId   string `survery:"redmineId"`
	Name        string `survery:"name"`
	ProjectRole string `survery:"projectRole"`
}

type SetupNumberMember struct {
	Number string `survey:"numberMember"`
}

type SetupNotice struct {
	TimeSendReport string `survey:"redmindSendReport"`
}

type Setup struct{}

func SurveyRun(file string) {
	setup := Setup{}
	setup.CopyFile(configs.PATH_GOOGLE_OAUTH_TMPL, configs.PATH_GOOGLE_OAUTH)
	setup.CopyFile(configs.PATH_GITHUB_OAUTH_TMPL, configs.PATH_GITHUB_OAUTH)
	setup.CopyFile(configs.PATH_CHATWORK_LOC_TEMPLATE_TMPL, configs.PATH_CHATWORK_LOC_TEMPLATE)
	setup.CopyFile(configs.PATH_CHATWORK_NOTIFY_INPROGRESS_REDMINE_TEMPLATE_TMPL, configs.PATH_CHATWORK_NOTIFY_INPROGRESS_REDMINE_TEMPLATE)
	setup.CopyFile(configs.PATH_CHATWORK_REDMINE_TEMPLATE_TMPL, configs.PATH_CHATWORK_REDMINE_TEMPLATE)
	setup.CopyFile(configs.PATH_REMIND_SEND_REPORT_TMPL, configs.PATH_REMIND_SEND_REPORT_TEMPLATE)

	switch file {
	case configs.ACTION_SETUP_GITHUB:
		fmt.Println("Setup github_secret.json")
		setup.SetupGithub()
	case configs.ACTION_SETUP_CHATWORK:
		fmt.Println("Setup chatwork_secret.json")
		setup.SetupChatwork()
	case configs.ACTION_SETUP_GOOGLE:
		fmt.Println("Setup google_secret.json")
		setup.SetupGoogle()
	case configs.ACTION_SETUP_SPREAD_SHEETS:
		fmt.Println("Setup spread_sheet.json")
		setup.SetupSpreadSheets()
	case configs.ACTION_SETUP_REDMINE:
		fmt.Println("Setup redmine_secret.json")
		setup.SetupRedmine()
	case configs.ACTION_SETUP_MEMBER:
		fmt.Println("Setup members.json")
		setup.SetupMember()
	case configs.ACTION_SETUP_ALL:
		setup.SetupAll()
	}
}

func (s Setup) CopyFile(srcFilePath string, destFilePath string) {
	srcFile, err := os.Open(srcFilePath)
	checkError(err)
	defer srcFile.Close()

	destFile, err := os.Create(destFilePath)
	checkError(err)
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	checkError(err)
}

func (s Setup) SetupAll() {
	fmt.Println("1. Setup github_secret.json")
	s.SetupGithub()
	fmt.Printf("\n")
	fmt.Println("2. Setup google_secret.json")
	s.SetupGoogle()
	fmt.Printf("\n")
	fmt.Println("3. Setup chatwork_secret.json")
	s.SetupChatwork()
	fmt.Printf("\n")
	fmt.Println("4. Setup spread_sheet.json")
	s.SetupSpreadSheets()
	fmt.Printf("\n")
	fmt.Println("5. Setup redmine_secret.json")
	s.SetupRedmine()
	fmt.Printf("\n")
	fmt.Println("6. Setup members.json")
	s.SetupMember()
}

func (s Setup) writeFileSetup(data interface{}, inputFilePath string, outputFilePath string) {
	t, err := template.ParseFiles(inputFilePath)
	checkError(err)

	var body bytes.Buffer
	err = t.Execute(&body, data)
	checkError(err)

	fmt.Printf("/=========== Your setup file ==========/")
	fmt.Printf("\n")
	fmt.Printf(body.String())
	fmt.Printf("/======================================/")
	fmt.Printf("\n")

	err = ioutil.WriteFile(outputFilePath, body.Bytes(), 0644)
	checkError(err)
}

func (s Setup) SetupGithub() {
	github := SetupGithubSecret{}
	githubQs := s.newGithubQs()

	err := survey.Ask(githubQs, &github)
	checkError(err)

	s.writeFileSetup(github, configs.PATH_GITHUB_SECRET_TMPL, configs.PATH_GITHUB_SECRET)
}

func (s Setup) SetupGoogle() {
	google := SetupGoogleSecret{}
	googleQs := s.newGoogleQs()

	err := survey.Ask(googleQs, &google)
	checkError(err)

	s.writeFileSetup(google, configs.PATH_GOOGLE_SECRET_TMPL, configs.PATH_GOOGLE_SECRET)
}

func (s Setup) SetupChatwork() {
	chatwork := SetupChatworkSecret{}
	chatworkQs := s.newChatworkQs()

	err := survey.Ask(chatworkQs, &chatwork)
	checkError(err)

	s.writeFileSetup(chatwork, configs.PATH_CHATWORK_SECRET_TMPL, configs.PATH_CHATWORK_SECRET)
}

func (s Setup) SetupSpreadSheets() {
	numberSpread := SetupNumberSpread{}
	numberSpreadQs := s.newNumberSpreadQs()
	err := survey.Ask(numberSpreadQs, &numberSpread)
	checkError(err)

	loopNumberSheetQs, errLoop := strconv.Atoi(numberSpread.Number)
	checkError(errLoop)
	var spreadSheets []SetupSpreadSheetsSecret = make([]SetupSpreadSheetsSecret, 0)

	spShQs := make(map[int][]*survey.Question)

	for i := 0; i < loopNumberSheetQs; i++ {
		fmt.Printf("\n")
		fmt.Println("Sheet ", i+1)
		n := SetupSpreadSheetsSecret{}
		spShQs[i] = s.newSpreadSheetQs()
		err = survey.Ask(spShQs[i], &n)
		checkError(err)
		spreadSheets = append(spreadSheets, n)
	}

	s.writeFileSetup(spreadSheets, configs.PATH_SPREAD_SHEETS_TMPL, configs.PATH_SPREAD_SHEETS)
}

func (s Setup) SetupRedmine() {
	redmine := SetupRedmineSecret{}
	redmineQs := s.newRedmineQs()

	err := survey.Ask(redmineQs, &redmine)
	checkError(err)

	s.writeFileSetup(redmine, configs.PATH_REDMINE_SECRET_TMPL, configs.PATH_REDMINE_SECRET)
}

func (s Setup) SetupMember() {
	numberMember := SetupNumberMember{}
	numberMemberQs := s.newNumberMemberQs()
	err := survey.Ask(numberMemberQs, &numberMember)
	checkError(err)

	loopNumberMemberQs, errLoop := strconv.Atoi(numberMember.Number)
	checkError(errLoop)

	var member []SetupMember = make([]SetupMember, 0)

	mbQs := make(map[int][]*survey.Question)

	for i := 0; i < loopNumberMemberQs; i++ {
		fmt.Println("\n")
		fmt.Println("Member ", i+1)
		n := SetupMember{}
		mbQs[i] = s.newMemberQs()
		err = survey.Ask(mbQs[i], &n)
		checkError(err)
		member = append(member, n)
	}

	s.writeFileSetup(member, configs.PATH_MEMBER_TMPL, configs.PATH_MEMBER)
}

func (s Setup) newSpreadSheetQs() []*survey.Question {
	var spreadSheetsQs = []*survey.Question{
		{
			Name: "id",
			Prompt: &survey.Input{
				Message: "What is spread sheets id?",
			},
			Validate: survey.Required,
		},
		{
			Name: "sheetLoc",
			Prompt: &survey.Input{
				Message: "What is column LOC of spreadsheet?",
			},
			Validate: survey.Required,
		},
		{
			Name: "cTicket",
			Prompt: &survey.Input{
				Message: "What is column Ticket?",
			},
			Validate: survey.Required,
		},
		{
			Name: "cGithub",
			Prompt: &survey.Input{
				Message: "What is column Github?",
			},
			Validate: survey.Required,
		},
		{
			Name: "cPoint",
			Prompt: &survey.Input{
				Message: "What is column Point?",
			},
			Validate: survey.Required,
		},
		{
			Name: "cLoc",
			Prompt: &survey.Input{
				Message: "What is column Loc?",
			},
			Validate: survey.Required,
		},
		{
			Name: "cRowStart",
			Prompt: &survey.Input{
				Message: "What is column Row start?",
			},
			Validate: survey.Required,
		},
	}
	return spreadSheetsQs
}

func (s Setup) newGithubQs() []*survey.Question {
	var githubQs = []*survey.Question{
		{
			Name: "githubClientId",
			Prompt: &survey.Input{
				Message: "What is your Github client id?",
			},
			Validate: survey.Required,
		},
		{
			Name: "githubClientSecret",
			Prompt: &survey.Input{
				Message: "What is your Github client secret?",
			},
			Validate: survey.Required,
		},
	}
	return githubQs
}

func (s Setup) newChatworkQs() []*survey.Question {
	var chatworkQs = []*survey.Question{
		{
			Name: "chatworkToken",
			Prompt: &survey.Input{
				Message: "What is your Chatwork token?",
			},
			Validate: survey.Required,
		},
		{
			Name: "chatworkRoomId",
			Prompt: &survey.Input{
				Message: "What is your Chatwork room id?",
			},
			Validate: survey.Required,
		},
	}
	return chatworkQs
}

func (s Setup) newGoogleQs() []*survey.Question {
	var googleQs = []*survey.Question{
		{
			Name: "googleClientId",
			Prompt: &survey.Input{
				Message: "What is your Google client id?",
			},
			Validate: survey.Required,
		},
		{
			Name: "googleClientSecret",
			Prompt: &survey.Input{
				Message: "What is your Google client secret?",
			},
			Validate: survey.Required,
		},
	}
	return googleQs
}

func (s Setup) newNumberSpreadQs() []*survey.Question {
	var numberSpreadQs = []*survey.Question{
		{
			Name: "numberSpreadQs",
			Prompt: &survey.Input{
				Message: "How many spread sheets you want to update?",
			},
			Validate: survey.Required,
		},
	}
	return numberSpreadQs
}

func (s Setup) newRedmineQs() []*survey.Question {
	var redmineQs = []*survey.Question{
		{
			Name: "redmineToken",
			Prompt: &survey.Input{
				Message: "What is your Redmine token?",
			},
			Validate: survey.Required,
		},
		{
			Name: "redmineUrl",
			Prompt: &survey.Input{
				Message: "What is your Redmine Url?",
			},
			Validate: survey.Required,
		},
		{
			Name: "redmineProjectIdentifier",
			Prompt: &survey.Input{
				Message: "What is your identifier Redmine project?",
			},
			Validate: survey.Required,
		},
	}
	return redmineQs
}

func (s Setup) newMemberQs() []*survey.Question {
	var memberQs = []*survey.Question{
		{
			Name: "chatworkId",
			Prompt: &survey.Input{
				Message: "What is chatwork id member?",
			},
			Validate: survey.Required,
		},
		{
			Name: "redmineId",
			Prompt: &survey.Input{
				Message: "What is redmine id member?",
			},
			Validate: survey.Required,
		},
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "What is member name?",
			},
			Validate: survey.Required,
		},
		{
			Name: "projectRole",
			Prompt: &survey.Select{
				Message: "What is member role",
				Options: []string{"Leader", "Developer", "Tester", "QA"},
			},
			Validate: survey.Required,
		},
	}
	return memberQs
}

func (s Setup) newNumberMemberQs() []*survey.Question {
	var numberMemberQs = []*survey.Question{
		{
			Name: "numberMember",
			Prompt: &survey.Input{
				Message: "How many member you want to update?",
			},
			Validate: survey.Required,
		},
	}
	return numberMemberQs
}

func (s Setup) newChatworkNotificationQs() []*survey.Question {
	var chatworkNoticeQs = []*survey.Question{
		{
			Name: "redmindSendReport",
			Prompt: &survey.Input{
				Message: "What time do you want to Redmine send report? (Ex: 07:06, 14:14)",
			},
			Validate: survey.Required,
		},
	}
	return chatworkNoticeQs
}

func (s Setup) SetupNotice() {
	notice := SetupNotice{}
	noticeQs := s.newChatworkNotificationQs()

	err := survey.Ask(noticeQs, &notice)
	checkError(err)
	convertTRSRToUTC, _ := time.Parse("15:04", notice.TimeSendReport)
	convertTRSRToUTC = convertTRSRToUTC.Add(-7 * time.Hour)
	hour, min, _ := convertTRSRToUTC.Clock()
	var strHour, strMin string
	if len(strconv.Itoa(hour)) == 1 {
		strHour = "0" + strconv.Itoa(hour)
	} else {
		strHour = strconv.Itoa(hour)
	}
	if len(strconv.Itoa(min)) == 1 {
		strMin = "0" + strconv.Itoa(min)
	} else {
		strMin = strconv.Itoa(min)
	}
	timeString := strHour + ":" + strMin
	SendNotice(timeString)
}
