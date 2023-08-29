package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New 创建一个新的 Mailer 实例，并将其作为函数返回。
func New(host string, port int, username, passwd, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, passwd)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

func (m Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// 执行 “subject” 模板，将其输出写入到 bytes.Buffer 中
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// 执行 “plainBody” 模板，将其输出写入到 bytes.Buffer 中
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// 执行 “htmlBody” 模板，将其输出写入到 bytes.Buffer 中
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// 使用 DialAndSend() 方法发送消息
	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}
	return nil
}
