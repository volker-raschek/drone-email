package mail

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"text/template"
	"time"

	"git.cryptic.systems/volker.raschek/drone-email-docker/pkg/domain"

	_ "embed"
)

const (
	DefaultSMTPFromAddress           = "root@localhost"
	DefaultSMTPFromName              = "root"
	DefaultSMTPHost                  = "localhost"
	DefaultSMTPPort                  = 587
	DefaultSMTPStartTLS              = true
	DefaultSMTPTLSInsecureSkipVerify = false
	DefaultSMTPToAddress             = "root@localhost"
)

//go:embed assets/mail.txt
var mailTemplate string

type CIVars struct {
	Build       *domain.Build
	Commit      *domain.Commit
	DeployTo    string
	Job         *domain.Job
	Prev        *domain.Prev
	PullRequest int
	Remote      *domain.Remote
	Repo        *domain.Repo
	Tag         string
	Yaml        *domain.Yaml
}

type templateVars struct {
	CIVars       *CIVars
	Recipient    string
	SMTPSettings *domain.SMTPSettings
}

func (t *templateVars) TimeNowFormat(layout string) string {
	return time.Now().Format(layout)
}

type Plugin struct {
	smtpSettings *domain.SMTPSettings
}

// Exec will send emails over SMTP
func (p *Plugin) Exec(ctx context.Context, recipients []string, ciVars *CIVars) error {
	exists := false
	for _, recipient := range recipients {
		if recipient == ciVars.Commit.Author.Email {
			exists = true
			break
		}
	}

	if !exists {
		recipients = append(recipients, ciVars.Commit.Author.Email)
	}

	tpl, err := template.New("mail").Parse(mailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	buf := make([]byte, 0)
	buffer := bytes.NewBuffer(buf)

	for _, recipient := range recipients {
		err = tpl.Execute(buffer, &templateVars{
			CIVars:       ciVars,
			Recipient:    recipient,
			SMTPSettings: p.smtpSettings,
		})
		if err != nil {
			return fmt.Errorf("failed to generate template: %w", err)
		}

		err := p.sendMail(recipient, buffer)
		if err != nil {
			return fmt.Errorf("failed to send mail: %w", err)
		}

		buffer.Reset()
	}

	return nil
}

func (p *Plugin) sendMail(recipient string, r io.Reader) error {
	// log.Printf("FROM_ADDRESS: %s", p.smtpSettings.FromAddress)
	// log.Printf("FROM_NAME: %s", p.smtpSettings.FromName)
	// log.Printf("HELO: %s", p.smtpSettings.HELOName)
	// log.Printf("HOST: %s", p.smtpSettings.Host)
	// log.Printf("PASSWORD: %s", p.smtpSettings.Password)
	// log.Printf("USERNAME: %s", p.smtpSettings.Username)
	// log.Printf("PORT: %v", p.smtpSettings.Port)
	// log.Printf("START_TLS: %v", p.smtpSettings.StartTLS)
	// log.Printf("INSECURE: %v", p.smtpSettings.TLSInsecureSkipVerify)

	address := fmt.Sprintf("%s:%d", p.smtpSettings.Host, p.smtpSettings.Port)
	tcpConn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to dial a connection to %s: %w", address, err)
	}
	defer func() { _ = tcpConn.Close() }()

	smtpClient, err := smtp.NewClient(tcpConn, p.smtpSettings.Host)
	if err != nil {
		return fmt.Errorf("failed to initialize a new smtp client: %w", err)
	}
	defer func() { _ = smtpClient.Close() }()

	err = smtpClient.Hello(p.smtpSettings.HELOName)
	if err != nil {
		return fmt.Errorf("failed to send helo command: %w", err)
	}

	// #nosec G402
	err = smtpClient.StartTLS(&tls.Config{
		InsecureSkipVerify: p.smtpSettings.TLSInsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
		ServerName:         p.smtpSettings.Host,
	})
	if err != nil {
		return fmt.Errorf("failed initialize starttls session: %w", err)
	}

	smtpAuth := smtp.PlainAuth(p.smtpSettings.FromAddress, p.smtpSettings.FromAddress, p.smtpSettings.Password, p.smtpSettings.Host)
	err = smtpClient.Auth(smtpAuth)
	if err != nil {
		return fmt.Errorf("failed to authenticate client: %w", err)
	}

	err = smtpClient.Mail(p.smtpSettings.FromAddress)
	if err != nil {
		return fmt.Errorf("failed to sent mail command: %w", err)
	}

	err = smtpClient.Rcpt(recipient)
	if err != nil {
		return fmt.Errorf("failed to sent rcpt command for %s: %w", recipient, err)
	}

	wc, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("failed to send data command: %w", err)
	}
	defer func() { _ = wc.Close() }()

	_, err = io.Copy(wc, r)
	if err != nil {
		return fmt.Errorf("failed to copy input from passed reader to smtp writer: %w", err)
	}

	// close smtpClient before defer to avoid returning an error of
	// smtpClient.Quit() like the following example:
	// Error: failed to execute mail plugin: failed to send mail: failed to send quit command: 250 2.0.0 Ok: queued as C7F009B4ED
	err = wc.Close()
	if err != nil {
		return fmt.Errorf("failed to close smtp client connection: %w", err)
	}

	err = smtpClient.Quit()
	if err != nil {
		return fmt.Errorf("failed to send quit command: %w", err)
	}

	return nil
}

func NewPlugin(config *domain.SMTPSettings) *Plugin {
	return &Plugin{
		smtpSettings: config,
	}
}
