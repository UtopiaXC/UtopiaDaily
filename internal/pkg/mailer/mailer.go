package mailer

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Enabled    bool
	Host       string
	Port       int
	From       string
	FromName   string
	User       string
	Password   string
	Encryption string
}

type Mailer struct {
	cfg Config
}

func NewMailer(cfg Config) *Mailer {
	return &Mailer{cfg: cfg}
}

type loginAuth struct {
	username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		prompt := strings.ToLower(string(fromServer))
		if strings.Contains(prompt, "username") {
			return []byte(a.username), nil
		}
		if strings.Contains(prompt, "password") {
			return []byte(a.password), nil
		}
		return nil, fmt.Errorf("unexpected SMTP auth prompt: %s", string(fromServer))
	}
	return nil, nil
}

func (m *Mailer) ValidateHost() error {
	if m.cfg.Host == "" {
		return errors.New("SMTP host is empty")
	}
	ips, err := net.LookupIP(m.cfg.Host)
	if err != nil {
		return fmt.Errorf("resolve SMTP host: %w", err)
	}
	for _, ip := range ips {
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("SMTP host resolves to restricted network address: %s", ip.String())
		}
	}
	return nil
}

func (m *Mailer) dial() (*smtp.Client, error) {
	if err := m.ValidateHost(); err != nil {
		return nil, err
	}
	port := m.cfg.Port
	if port <= 0 {
		switch strings.ToLower(m.cfg.Encryption) {
		case "ssl":
			port = 465
		case "tls":
			port = 587
		default:
			port = 25
		}
	}

	addr := net.JoinHostPort(m.cfg.Host, strconv.Itoa(port))
	tlsCfg := &tls.Config{ServerName: m.cfg.Host, MinVersion: tls.VersionTLS12}
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	var client *smtp.Client
	var err error
	if strings.EqualFold(m.cfg.Encryption, "ssl") || (m.cfg.Encryption == "" && port == 465) {
		conn, dialErr := tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
		if dialErr != nil {
			return nil, fmt.Errorf("connect SMTP SSL: %w", dialErr)
		}
		client, err = smtp.NewClient(conn, m.cfg.Host)
		if err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("create SMTP client: %w", err)
		}
	} else {
		conn, dialErr := dialer.Dial("tcp", addr)
		if dialErr != nil {
			return nil, fmt.Errorf("connect SMTP: %w", dialErr)
		}
		client, err = smtp.NewClient(conn, m.cfg.Host)
		if err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("create SMTP client: %w", err)
		}
		if strings.EqualFold(m.cfg.Encryption, "tls") || port == 587 {
			if ok, _ := client.Extension("STARTTLS"); ok {
				if err := client.StartTLS(tlsCfg); err != nil {
					_ = client.Close()
					return nil, fmt.Errorf("SMTP STARTTLS: %w", err)
				}
			}
		}
	}

	if m.cfg.User != "" {
		auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Password, m.cfg.Host)
		if err := client.Auth(auth); err != nil {
			login := &loginAuth{username: m.cfg.User, password: m.cfg.Password}
			if loginErr := client.Auth(login); loginErr != nil {
				_ = client.Close()
				return nil, fmt.Errorf("SMTP auth failed: %w", err)
			}
		}
	}

	return client, nil
}

func (m *Mailer) TestConnection() error {
	if m.cfg.Host == "" {
		return errors.New("SMTP host not configured")
	}
	client, err := m.dial()
	if err != nil {
		return err
	}
	return client.Quit()
}

func (m *Mailer) SendHTML(to, subject, htmlBody, textBody string) error {
	if !m.cfg.Enabled {
		return errors.New("SMTP disabled")
	}
	if m.cfg.Host == "" || m.cfg.From == "" {
		return errors.New("SMTP host or sender not configured")
	}
	to = sanitizeHeader(to)
	if to == "" {
		return errors.New("recipient address required")
	}

	client, err := m.dial()
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	fromAddr := sanitizeHeader(m.cfg.From)
	fromName := sanitizeHeader(m.cfg.FromName)
	if fromName == "" {
		fromName = "Utopia Daily"
	}

	if err := client.Mail(fromAddr); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}

	boundary := randomBoundary()
	encodedFrom := fmt.Sprintf("=?UTF-8?B?%s?= <%s>", base64.StdEncoding.EncodeToString([]byte(fromName)), fromAddr)
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(sanitizeHeader(subject))))

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", encodedFrom))
	msg.WriteString(fmt.Sprintf("To: <%s>\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

	if textBody != "" {
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
		msg.WriteString(wrapBase64([]byte(textBody)) + "\r\n")
	}

	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	msg.WriteString(wrapBase64([]byte(htmlBody)) + "\r\n")
	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	if _, err := w.Write([]byte(msg.String())); err != nil {
		return fmt.Errorf("write email body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close email stream: %w", err)
	}
	return client.Quit()
}

func wrapBase64(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	var b strings.Builder
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		b.WriteString(encoded[i:end])
		b.WriteString("\r\n")
	}
	return b.String()
}

func randomBoundary() string {
	buf := make([]byte, 12)
	_, _ = rand.Read(buf)
	return "utd_" + hex.EncodeToString(buf)
}

func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return strings.TrimSpace(s)
}
