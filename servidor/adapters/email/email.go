// Package email envia mensagens por SMTP. Sem SMTP configurado, entra em modo de teste:
// a mensagem não sai e o texto vai só para o log do servidor (use apenas no computador local).
package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/environment"
)

type Email_I interface {
	Enviar(para, assunto, texto string) error
}

type smtp_t struct{ cfg environment.EmailConfig_t }

type teste_t struct{}

var (
	local     Email_I
	localOnce sync.Once
)

// NewEmail devolve o envio único configurado pelo environment
func NewEmail() Email_I {
	localOnce.Do(func() {
		cfg := environment.Cfg.Email
		if cfg.Configurado() {
			local = &smtp_t{cfg: cfg}
			return
		}
		logrus.Warn("email: SMTP não configurado; modo de teste (os códigos aparecem só neste log)")
		local = &teste_t{}
	})
	return local
}

func (t *teste_t) Enviar(para, assunto, texto string) error {
	logrus.Warnf("email (modo de teste, não enviado) para %s: %s\n%s", para, assunto, texto)
	return nil
}

func (s *smtp_t) Enviar(para, assunto, texto string) error {
	if strings.ContainsAny(para, "\r\n") {
		return fmt.Errorf("email: destinatário inválido")
	}
	var msg bytes.Buffer
	fmt.Fprintf(&msg, "From: %s\r\n", mime.QEncoding.Encode("utf-8", "Retenção Gobrax")+" <"+s.cfg.SMTPRemetente+">")
	fmt.Fprintf(&msg, "To: %s\r\n", para)
	fmt.Fprintf(&msg, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", assunto))
	fmt.Fprintf(&msg, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	msg.WriteString("MIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n")
	msg.WriteString(strings.ReplaceAll(texto, "\n", "\r\n"))

	end := net.JoinHostPort(s.cfg.SMTPHost, strconv.Itoa(s.cfg.SMTPPorta))
	var auth smtp.Auth
	if s.cfg.SMTPUsuario != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUsuario, s.cfg.SMTPSenha, s.cfg.SMTPHost)
	}
	if s.cfg.SMTPPorta == 465 { // TLS direto
		conn, err := tls.Dial("tcp", end, &tls.Config{ServerName: s.cfg.SMTPHost, MinVersion: tls.VersionTLS12})
		if err != nil {
			return fmt.Errorf("email: conexão: %w", err)
		}
		c, err := smtp.NewClient(conn, s.cfg.SMTPHost)
		if err != nil {
			return err
		}
		defer c.Close()
		if auth != nil {
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("email: login no SMTP: %w", err)
			}
		}
		if err := c.Mail(s.cfg.SMTPRemetente); err != nil {
			return err
		}
		if err := c.Rcpt(para); err != nil {
			return err
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write(msg.Bytes()); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
		return c.Quit()
	}
	// 587/25: smtp.SendMail usa STARTTLS quando o servidor oferece
	return smtp.SendMail(end, auth, s.cfg.SMTPRemetente, []string{para}, msg.Bytes())
}
