// Package twilio envia mensagens de WhatsApp por template (Content API da Twilio).
// Sem a conta configurada, entra em modo de teste: nada sai e o texto vai só para o log (com o telefone mascarado).
package twilio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/environment"
)

type WhatsApp_I interface {
	// Enviar manda o template configurado para o telefone (+55...) com as variáveis {"1": ..., "2": ...}; devolve o SID da mensagem.
	// Sem template configurado, manda o texto pronto: só chega a quem falou com o número nas últimas 24 h (Sandbox, testes).
	Enviar(ctx context.Context, telefone string, variaveis map[string]string, texto string) (string, error)
	// Teste indica o modo de teste (nada é enviado de verdade)
	Teste() bool
}

// Twilio_t chama a API de verdade; Base e HTTP são trocados nos testes
type Twilio_t struct {
	Cfg  environment.WhatsAppConfig_t
	Base string
	HTTP *http.Client
}

type teste_t struct{}

const baseTwilio = "https://api.twilio.com/2010-04-01/Accounts/"

var clienteHTTP = &http.Client{Timeout: 20 * time.Second}

// NewWhatsApp devolve o envio para a configuração dada (a da aba WhatsApp ou a do environment);
// incompleta, cai no modo de teste. Com o Auth Token, TwilioAPIKey é o próprio Account SID.
func NewWhatsApp(cfg environment.WhatsAppConfig_t) WhatsApp_I {
	if cfg.Configurado() {
		return &Twilio_t{Cfg: cfg, Base: baseTwilio, HTTP: clienteHTTP}
	}
	return &teste_t{}
}

func (t *teste_t) Teste() bool { return true }

func (t *teste_t) Enviar(_ context.Context, telefone string, variaveis map[string]string, _ string) (string, error) {
	logrus.Warnf("whatsapp (modo de teste, não enviado) para %s: %v", Mascarar(telefone), variaveis)
	return "", nil
}

func (t *Twilio_t) Teste() bool { return false }

// erroTwilio é o corpo que a Twilio devolve quando recusa a mensagem
type erroTwilio struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (t *Twilio_t) Enviar(ctx context.Context, telefone string, variaveis map[string]string, texto string) (string, error) {
	form := url.Values{"From": {t.Cfg.Remetente}, "To": {"whatsapp:" + telefone}}
	if t.Cfg.TemplateSID == "" {
		form.Set("Body", texto)
	} else {
		vars, err := json.Marshal(variaveis)
		if err != nil {
			return "", err
		}
		form.Set("ContentSid", t.Cfg.TemplateSID)
		form.Set("ContentVariables", string(vars))
	}
	end := t.Base + url.PathEscape(t.Cfg.TwilioAccountSID) + "/Messages.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, end, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(t.Cfg.TwilioAPIKey, t.Cfg.TwilioAPISecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("twilio: conexão: %w", err)
	}
	defer resp.Body.Close()
	corpo, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrCredenciais
	}
	if resp.StatusCode >= 300 {
		var e erroTwilio
		_ = json.Unmarshal(corpo, &e)
		if e.Message == "" {
			e.Message = http.StatusText(resp.StatusCode)
		}
		return "", fmt.Errorf("twilio recusou (código %d): %s", e.Code, e.Message)
	}
	var ok struct {
		SID string `json:"sid"`
	}
	if err := json.Unmarshal(corpo, &ok); err != nil || ok.SID == "" {
		return "", errors.New("twilio: resposta sem o SID da mensagem")
	}
	return ok.SID, nil
}

// ErrCredenciais: a API Key ou o segredo não valem (revogados ou copiados errado). Ninguém recebe até corrigir.
var ErrCredenciais = errors.New("a Twilio recusou o Account SID e o Auth Token; confira os dois na Conexão com a Twilio (aba WhatsApp)")

// Mascarar deixa só os 4 últimos dígitos do telefone, para log e tela
func Mascarar(telefone string) string {
	if len(telefone) < 4 {
		return "****"
	}
	return "••••" + telefone[len(telefone)-4:]
}
