package whatsapp

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"analisefrota/common/cripto"
	"analisefrota/environment"
	"analisefrota/modules/whatsapp/whatsapp_repo"
)

// Config_t é a conexão com a Twilio como a tela vê: o Auth Token nunca sai, só se ele existe
type Config_t struct {
	AccountSID   string     `json:"account_sid"`
	TemToken     bool       `json:"tem_token"`
	Remetente    string     `json:"remetente"`
	TemplateSID  string     `json:"template_sid"`
	Automatico   bool       `json:"automatico"`
	Hora         int        `json:"hora"`
	Configurado  bool       `json:"configurado"`
	Origem       string     `json:"origem"` // "tela", "ambiente" ou "nenhuma"
	AtualizadoEm *time.Time `json:"atualizado_em"`
}

// NovaConfig_t é o que o administrador salva; AuthToken vazio mantém o que já estava salvo
type NovaConfig_t struct {
	AccountSID  string
	AuthToken   string
	Remetente   string
	TemplateSID string
	Automatico  bool
	Hora        int
}

// remetente padrão do Sandbox da Twilio (para testar sem número aprovado)
const RemetenteSandbox = "whatsapp:+14155238886"

var (
	ErrAccountSID  = errors.New("Account SID inválido: começa com AC e tem 34 caracteres (copie do console da Twilio)")
	ErrAuthToken   = errors.New("Auth Token inválido: tem 32 letras e números (copie do console da Twilio)")
	ErrSemToken    = errors.New("informe o Auth Token da Twilio")
	ErrRemetente   = errors.New("número que envia inválido: use o número do WhatsApp da Twilio com código do país, por exemplo +1 415 523 8886")
	ErrTemplateSID = errors.New("Template SID inválido: começa com HX e tem 34 caracteres (Content Template Builder)")
	ErrHora        = errors.New("a hora do envio vai de 0 a 23")

	reAccountSID  = regexp.MustCompile(`^AC[0-9a-fA-F]{32}$`)
	reAuthToken   = regexp.MustCompile(`^[0-9a-zA-Z]{32}$`)
	reTemplateSID = regexp.MustCompile(`^HX[0-9a-fA-F]{32}$`)
	naoDigito     = regexp.MustCompile(`\D`)
)

// normalizarRemetente aceita "+1 415 523 8886", "14155238886" ou "whatsapp:+14155238886"
func normalizarRemetente(v string) (string, error) {
	d := naoDigito.ReplaceAllString(strings.TrimPrefix(strings.TrimSpace(v), "whatsapp:"), "")
	if len(d) < 10 || len(d) > 15 {
		return "", ErrRemetente
	}
	return "whatsapp:+" + d, nil
}

func (w *whatsapp_t) Config(ctx context.Context) (Config_t, error) {
	row, ok, err := w.d.Repo.BuscarConfig(ctx)
	if err != nil {
		return Config_t{}, err
	}
	if ok {
		em := row.AtualizadoEm
		return Config_t{AccountSID: row.AccountSID, TemToken: len(row.AuthTokenCifrado) > 0, Remetente: row.Remetente,
			TemplateSID: row.TemplateSID, Automatico: row.Automatico, Hora: row.Hora, Configurado: true, Origem: "tela", AtualizadoEm: &em}, nil
	}
	env := w.d.Env
	origem := "nenhuma"
	if env.Configurado() {
		origem = "ambiente"
	}
	return Config_t{AccountSID: env.TwilioAccountSID, TemToken: env.TwilioAPISecret != "", Remetente: env.Remetente, TemplateSID: env.TemplateSID,
		Automatico: env.Automatico, Hora: env.Hora, Configurado: env.Configurado(), Origem: origem}, nil
}

func (w *whatsapp_t) SalvarConfig(ctx context.Context, n NovaConfig_t, usuarioID int64) (Config_t, error) {
	n.AccountSID, n.TemplateSID = strings.TrimSpace(n.AccountSID), strings.TrimSpace(n.TemplateSID)
	n.AuthToken = strings.TrimSpace(n.AuthToken)
	if !reAccountSID.MatchString(n.AccountSID) {
		return Config_t{}, ErrAccountSID
	}
	if !reTemplateSID.MatchString(n.TemplateSID) {
		return Config_t{}, ErrTemplateSID
	}
	rem, err := normalizarRemetente(n.Remetente)
	if err != nil {
		return Config_t{}, err
	}
	if n.Hora < 0 || n.Hora > 23 {
		return Config_t{}, ErrHora
	}
	atual, existe, err := w.d.Repo.BuscarConfig(ctx)
	if err != nil {
		return Config_t{}, err
	}
	row := whatsapp_repo.ConfigRow_t{AccountSID: n.AccountSID, Remetente: rem, TemplateSID: n.TemplateSID, Automatico: n.Automatico, Hora: n.Hora}
	switch {
	case n.AuthToken != "":
		if !reAuthToken.MatchString(n.AuthToken) {
			return Config_t{}, ErrAuthToken
		}
		if row.AuthTokenCifrado, err = cripto.Cifrar(w.d.Chave, n.AuthToken); err != nil {
			return Config_t{}, err
		}
	case existe:
		row.AuthTokenCifrado = atual.AuthTokenCifrado
	default:
		return Config_t{}, ErrSemToken
	}
	if usuarioID > 0 {
		row.AtualizadoPor = &usuarioID
	}
	if err := w.d.Repo.SalvarConfig(ctx, row); err != nil {
		return Config_t{}, err
	}
	return w.Config(ctx)
}

// efetiva é a configuração que vale para enviar: a da tela, se o administrador salvou; senão, a do ambiente
func (w *whatsapp_t) efetiva(ctx context.Context) (environment.WhatsAppConfig_t, error) {
	row, ok, err := w.d.Repo.BuscarConfig(ctx)
	if err != nil || !ok {
		return w.d.Env, err
	}
	token, err := cripto.Decifrar(w.d.Chave, row.AuthTokenCifrado)
	if err != nil {
		return environment.WhatsAppConfig_t{}, errors.New("o Auth Token salvo não abre (CHAVE_TOKENS trocada?); salve de novo na Conexão com a Twilio")
	}
	return environment.WhatsAppConfig_t{TwilioAccountSID: row.AccountSID, TwilioAPIKey: row.AccountSID, TwilioAPISecret: token,
		Remetente: row.Remetente, TemplateSID: row.TemplateSID, Automatico: row.Automatico, Hora: row.Hora}, nil
}
