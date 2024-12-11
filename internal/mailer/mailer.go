package mailer

import (
	"bytes"
	"embed"
	"github.com/wneessen/go-mail"
	"html/template"
	"time"
)

// templateFS is a variable with the type embed.FS to hold
// the email templates inside the templates directory.
//
//go:embed templates
var templateFS embed.FS

// Mailer is a struct which contains mail.Client instance(used to connect to an SMTP server)
// sender information for emails (the name and address you want the email to be from,
// such as "Alice Smith <alicesmith@example.com>".) and also methods for sending emails.
type Mailer struct {
	client *mail.Client
	sender string
}

// New returns a new Mailer with configured mail.Client and sender information.
func New(client *mail.Client, sender string) Mailer {
	return Mailer{client: client, sender: sender}
}

// Send is a method that send an email, with the provided template to the recipient.
//
// It takes the recipient email address as the first parameter,
// the name of the file containing the templates, and any dynamic data for
// the templates as any parameter.
func (m Mailer) Send(recipient, templateFile string, data any) error {
	// Use the ParseFS() to parse the required template file from
	// the embedded file system.
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and storing
	// the result in a bytes.Buffer.
	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return err
	}

	// Execute the named template "plainBody", passing in the dynamic data and storing
	// the result in a bytes.Buffer.
	plainBody := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(plainBody, "plainBody", data); err != nil {
		return err
	}

	// Execute the named template "html", passing in the dynamic data and storing
	// the result in a bytes.Buffer.
	htmlBody := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(htmlBody, "htmlBody", data); err != nil {
		return err
	}

	// Use the mail.NewMsg() to initialize a new mail.Message instance.
	// Then set the address "TO" and "FROM", and also the plain-text body
	// and html body alternative.
	msg := mail.NewMsg()
	msg.SetGenHeader(mail.HeaderSubject, subject.String())
	if err := msg.SetAddrHeader(mail.HeaderTo, recipient); err != nil {
		return err
	}
	if err := msg.SetAddrHeader(mail.HeaderFrom, m.sender); err != nil {
		return err
	}
	msg.SetBodyString(mail.TypeTextPlain, plainBody.String())
	msg.AddAlternativeString(mail.TypeTextHTML, htmlBody.String())

	// Send the message with the maximum of 3 retries
	// sleep for a second between each attempt.
	for i := 0; i < 3; i++ {
		if err = m.client.Send(msg); nil == err {
			return nil
		}
		time.Sleep(30 * time.Second)
	}
	return err
}
