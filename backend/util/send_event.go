package util

import (
	"log"
	"net/smtp"

	"github.com/domodwyer/mailyak"
)

func init() {
	// 1 worker for sending email synchronously
	go sendEvents()
}

var (
	eventJobs = make(chan string, 50)
)

func SendEvent(msg string) {
	eventJobs <- msg
}

func sendEvents() {

	for event := range eventJobs {
		if Config.Production {
			mail := mailyak.New(Config.SMTPHost+":"+Config.SMTPPort, smtp.PlainAuth(Config.SMTPUser, Config.SMTPUser, Config.SMTPPass, Config.SMTPHost))
			mail.Plain().Set(event)
			mail.Subject("Monero Pot Event")
			mail.To(Config.ContactEmail)
			mail.From(Config.SMTPUser)
			if err := mail.Send(); err != nil {
				log.Println("Failed send event", event)
				eventJobs <- event
			}
		} else {
			log.Println("SendEvent", event)
		}
	}

}
