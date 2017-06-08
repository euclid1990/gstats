package utilities

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/euclid1990/gstats/configs"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const BaseUrl = `https://api.chatwork.com/v2`

type ChatworkConfig struct {
	CWToken  string `json:"token"`
	CWRoomId string `json:"room_id"`
}

type Chatwork struct {
	config  *ChatworkConfig
	tmpl    string
	BaseUrl string
}

type ChatworkSendMessageResponse struct {
	MessageId string `json:"message_id"`
}

func NewChatwork() *Chatwork {
	chatwork := new(Chatwork)
	chatwork.readConfig()
	chatwork.BaseUrl = BaseUrl
	return chatwork
}

func (c *Chatwork) readConfig() {
	var config ChatworkConfig
	b, err := ioutil.ReadFile(configs.PATH_CHATWORK_SECRET)
	if err != nil {
		log.Fatalf("[Chatwork] Unable to read client secret file: %v", err)
	}
	if err = json.Unmarshal(b, &config); err != nil {
		log.Fatalf("[Chatwork] Unable to parse client secret file to config: %v", err)
	}
	c.config = &config

}

func (c *Chatwork) setTemplate(templateName string) {
	c.tmpl = templateName
}

func (c *Chatwork) buildBody(params map[string]string) url.Values {
	body := url.Values{}
	for k := range params {
		body.Add(k, params[k])
	}
	return body
}

func (c *Chatwork) sendMessage(endpoint string, params map[string]string) error {
	var result ChatworkSendMessageResponse
	client := &http.Client{}
	req, requestErr := http.NewRequest("POST", c.BaseUrl+endpoint, bytes.NewBufferString(c.buildBody(params).Encode()))
	if requestErr != nil {
		return requestErr
	}
	req.Header.Add("X-ChatWorkToken", c.config.CWToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return err
	}
	if result.MessageId == "" {
		return errors.New("[Chatwork] Sending message failed")
	}
	return nil
}

func SendLocMessage(loc []Loc) error {
	var body bytes.Buffer
	chatwork := NewChatwork()
	chatwork.setTemplate(configs.PATH_CHATWORK_LOC_TEMPLATE)

	t := template.Must(template.New(strings.Split(configs.PATH_CHATWORK_LOC_TEMPLATE, "/")[1]).ParseFiles(chatwork.tmpl))
	err := t.Execute(&body, loc)
	if err != nil {
		panic(err)
	}
	sendErr := chatwork.sendMessage("/rooms/"+chatwork.config.CWRoomId+"/messages", map[string]string{"body": body.String()})
	if sendErr != nil {
		return sendErr
	}
	return nil
}

func (c *Chatwork) SendInprogressIssuesMessage(data []redmineNotify) {
	chatwork := NewChatwork()
	chatwork.setTemplate(configs.PATH_CHATWORK_REDMINE_TEMPLATE)

	var body bytes.Buffer
	t := template.Must(template.New(strings.Split(configs.PATH_CHATWORK_REDMINE_TEMPLATE, "/")[1]).ParseFiles(chatwork.tmpl))
	exeErr := t.Execute(&body, data)
	checkErrThrowLog(exeErr)
	err := chatwork.sendMessage("/rooms/"+chatwork.config.CWRoomId+"/messages", map[string]string{
		"body": body.String(),
	})
	checkErrThrowLog(err)
}

func (c *Chatwork) SendNoticeMemberNotHaveTaskInprogeress(data redmineNoticeUser) {
	var body bytes.Buffer
	chatwork := NewChatwork()
	chatwork.setTemplate(configs.PATH_REDMINE_NOTICE_USER)

	t := template.Must(template.New(strings.Split(configs.PATH_REDMINE_NOTICE_USER, "/")[1]).ParseFiles(chatwork.tmpl))
	err := t.Execute(&body, data)
	checkErrThrowLog(err)
	sendErr := chatwork.sendMessage("/rooms/"+chatwork.config.CWRoomId+"/messages", map[string]string{"body": body.String()})
	checkErrThrowLog(sendErr)
}
