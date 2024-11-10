package mails

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	Dialer       *mail.Dialer
	Sender       string
	RetriesCount int
}

func New(host string, port int, timeout time.Duration, username, password, sender string, retriesCount int) *Mailer {
	dialer := mail.NewDialer(host, port, sender, password)
	dialer.Timeout = timeout
	return &Mailer{
		Dialer:       dialer,
		Sender:       sender,
		RetriesCount: retriesCount,
	}
}

func parseEmailTmpl(tmplName string, tmplData any) (map[string]string, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/"+tmplName)
	if err != nil {
		return nil, err
	}
	tmplPartials := map[string]string{
		"subject":   "",
		"plainBody": "",
		"htmlBody":  "",
	}
	for key := range tmplPartials {
		buff := new(bytes.Buffer)
		if err = tmpl.ExecuteTemplate(buff, key, tmplData); err != nil {
			return nil, err
		}
		tmplPartials[key] = buff.String()
	}
	return tmplPartials, nil
}

func (m *Mailer) Send(recipient string, tmplName string, tmplData any) error {
	tmplPartials, err := parseEmailTmpl(tmplName, tmplData)
	if err != nil {
		return err
	}
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.Sender)
	msg.SetHeader("Subject", tmplPartials["subject"])
	msg.SetBody("text/plain", tmplPartials["plainBody"])
	msg.SetBody("text/html", tmplPartials["htmlBody"])
	for i := 0; i < m.RetriesCount; i++ {
		err = m.Dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

type ApiMailer struct {
	ApiToken     string
	Sender       string
	RetriesCount int
}

func (m *ApiMailer) Send(recipient string, tmplName string, tmplData any) error {
	const apiUrl = "https://sandbox.api.mailtrap.io/api/send/3112947"
	tmplPartials, err := parseEmailTmpl(tmplName, tmplData)
	if err != nil {
		return err
	}
	sender := strings.Split(m.Sender, " ")
	payload, err := json.Marshal(map[string]any{
		"from":    map[string]string{"email": sender[1], "name": sender[0]},
		"to":      []map[string]string{{"email": recipient}},
		"subject": tmplPartials["subject"],
		"text":    tmplPartials["plainBody"],
		"html":    tmplPartials["htmlBody"],
	})
	if err != nil {
		return err
	}
	payloadReader := strings.NewReader(string(payload))
	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, apiUrl, payloadReader)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+m.ApiToken)
	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	for i := 0; i < m.RetriesCount; i++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		return err
	}
	if resp != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var bodyParsed map[string]any
		err = json.Unmarshal(body, &bodyParsed)
		if err == nil {
			_, ok := bodyParsed["errors"]
			if ok {
				return fmt.Errorf("failed to send email: %s", bodyParsed["errors"])
			}
		}
		fmt.Println(string(body))
		defer resp.Body.Close()
	}
	return nil
}
