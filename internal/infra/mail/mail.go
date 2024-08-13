package mail

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/9ssi7/bank/assets"
	smtp_mail "github.com/xhit/go-simple-mail/v2"
)

type Srv interface {
	SendText(ctx context.Context, cnf SendConfig) error
	SendWithTemplate(ctx context.Context, cnf SendWithTemplateConfig) error
}

type Config struct {
	Host     string
	Port     int
	Sender   string
	Password string
	From     string
	Reply    string
}

type SendConfig struct {
	To      []string
	Subject string
	Message string
}

type SendWithTemplateConfig struct {
	SendConfig
	Template string
	Data     any
}

type srv struct {
	cnf    Config
	server *smtp_mail.SMTPServer
}

func Init(cnf Config) Srv {
	server := smtp_mail.NewSMTPClient()
	server.Host = cnf.Host
	server.Port = cnf.Port
	server.Username = cnf.Sender
	server.Password = cnf.Password
	server.Encryption = smtp_mail.EncryptionSTARTTLS
	server.Authentication = smtp_mail.AuthLogin
	return &srv{
		server: server,
		cnf:    cnf,
	}
}

func GetField(str string) string {
	if str == "" {
		return "N/A"
	}
	return str
}

func (s *srv) createClient() (*smtp_mail.SMTPClient, error) {
	return s.server.Connect()
}

func (s *srv) SendText(ctx context.Context, cnf SendConfig) error {
	client, err := s.createClient()
	if err != nil {
		fmt.Println("Error creating client: ", err)
		return err
	}
	email := smtp_mail.NewMSG()
	email.SetFrom(s.cnf.From)
	email.AddTo(cnf.To...)
	email.SetSubject(cnf.Subject)
	email.SetSender(s.cnf.Sender)
	if s.cnf.Reply != "" {
		email.SetReplyTo(s.cnf.Reply)
	}
	email.AddAlternative(smtp_mail.TextPlain, cnf.Message)
	err = email.Send(client)
	if err != nil {
		fmt.Println("Error sending email: ", err)
		return err
	}
	return nil
}

func (s *srv) SendWithTemplate(ctx context.Context, cnf SendWithTemplateConfig) error {
	client, err := s.createClient()
	if err != nil {
		fmt.Println("Error creating client: ", err)
		return err
	}
	dir := assets.EmbedMailTemplate()
	t := template.Must(template.ParseFS(dir, fmt.Sprintf("mail/%s.html", cnf.Template)))
	var tpl bytes.Buffer
	t.Execute(&tpl, cnf.Data)
	body := tpl.String()
	email := smtp_mail.NewMSG()
	email.SetFrom(s.cnf.From)
	email.AddTo(cnf.To...)
	email.SetSubject(cnf.Subject)
	email.SetSender(s.cnf.Sender)
	email.SetReplyTo(s.cnf.Reply)
	email.SetBody(smtp_mail.TextHTML, body)
	if err = email.Send(client); err != nil {
		fmt.Println("Error sending email: ", err)
		return err
	}
	return nil
}
