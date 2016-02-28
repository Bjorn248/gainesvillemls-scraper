package main

import (
	"fmt"
	"github.com/BjornTwitchBot/gainesvillemls-scraper/Godeps/_workspace/src/github.com/sendgrid/sendgrid-go"
	"os"
)

func sendEmail(email string, links []string) {
	sg := sendgrid.NewSendGridClientWithApiKey(os.Getenv("SENDGRID_API_TOKEN"))
	message := sendgrid.NewMail()
	message.AddTo(email)
	message.SetSubject("New Gainesville MLS Listing(s)")
	emailBody := ""
	for _, link := range links {
		emailBody = emailBody + fmt.Sprintf("<a href=%s>%s</a><br>", link, link)
	}
	message.SetHTML(emailBody)
	message.SetFrom(os.Getenv("EMAIL_FROM_ADDRESS"))
	r := sg.Send(message)
	if r != nil {
		fmt.Printf("Error sending email: '%s'", r)
		return
	}
}
