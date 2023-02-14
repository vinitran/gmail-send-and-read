package main

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/go-gomail/gomail"
	"github.com/joho/godotenv"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	//sendMail()
	getMail()

}

func sendMail() {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("ACCOUNT"))
	m.SetHeader("To", "20020503@vnu.edu.vn")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", fmt.Sprintf("Hello %s!", "Vinh"))
	m.Attach("attachment/1")

	d := gomail.NewDialer(os.Getenv("SMTP_SERVER"), 587, os.Getenv("ACCOUNT"), os.Getenv("PASSWORD"))

	// Send the email to Bob, Cora and Dan.
	err := d.DialAndSend(m)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Send mail successfully")
}

func getMail() {
	log.Println("Connecting to server...")
	// Connect to server
	c, err := client.DialTLS(os.Getenv("IMAP_SERVER"), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	defer c.Logout()

	// Login
	err = c.Login(os.Getenv("ACCOUNT"), os.Getenv("PASSWORD"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(mbox.Messages, mbox.Messages)

	messages := make(chan *imap.Message, 10)
	var section imap.BodySectionName
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope, imap.FetchRFC822}, messages)
	}()

	for msg := range messages {
		if msg == nil {
			continue
		}

		bd := msg.GetBody(&section)
		if bd == nil {
			continue
		}

		body, err := mail.CreateReader(bd)
		if err != nil {
			fmt.Println(err)
			continue
		}

		subject := msg.Envelope.Subject
		subject = strings.Replace(subject, " ", "-", -1)
		subject = strings.Replace(subject, ".", "-", -1)
		subject = strings.Replace(subject, "/", "-", -1)
		text := textToPush(msg, body)
		path, err := CreateFolder(fmt.Sprintf("mail/%s", subject))
		if err != nil {
			fmt.Println("Create folder: ", err)
			continue
		}
		err = CreateFileAndWrite(fmt.Sprintf("%s/%s", path, subject), text)
		if err != nil {
			fmt.Println("Create and write file: ", err)
			continue
		}
		getAttachmentAndSave(body, path)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}

func textToPush(msg *imap.Message, bd *mail.Reader) string {
	subject := msg.Envelope.Subject
	from := msg.Envelope.From[0].Address()
	to := msg.Envelope.To[0].Address()
	date := msg.Envelope.Date

	body, err := getBodyMail(bd)
	if err != nil {
		fmt.Println(err)
	}

	text := fmt.Sprintf("Title: %s\nFrom: %s\nTo: %s\nDate: %s\n\n%s", subject, from, to, date, body)
	return text
}

func getBodyMail(body *mail.Reader) (string, error) {
	p, err := body.NextPart()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	b, err := io.ReadAll(p.Body)
	if err != nil {
		fmt.Println(err)
	}

	return string(b), nil
}

func getAttachmentAndSave(body *mail.Reader, path string) {
	for {
		p, err := body.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			continue
		}

		switch h := p.Header.(type) {
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			fmt.Printf("Got attachment: %s", filename)
			path := fmt.Sprintf("%s/%s", path, filename)

			file, _ := os.Create(path)
			_, err := io.Copy(file, p.Body)
			if err != nil {
				fmt.Println(err)
				continue
			}
		default:

		}
	}
}
