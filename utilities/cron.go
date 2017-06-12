package utilities

import (
	"github.com/jasonlvhit/gocron"
)

func sendRemindReport() {
	SendRemindDailyReport()
}

func SendNotice(timeSetup string) {
	gocron.Every(1).Day().At(timeSetup).Do(sendRemindReport)
	<-gocron.Start()
}
