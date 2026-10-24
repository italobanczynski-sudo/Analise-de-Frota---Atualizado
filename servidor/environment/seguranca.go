package environment

import (
	"encoding/base64"
	"errors"
)

type SegurancaConfig_t struct {
	// Chave AES-256 (32 bytes em base64) que cifra os tokens dos clientes no banco
	ChaveTokens string `envconfig:"CHAVE_TOKENS" required:"true"`
	// Quanto tempo o login dura e quanto tempo o link de convite vale
	SessaoHoras  int `envconfig:"SESSAO_HORAS" default:"12"`
	ConviteHoras int `envconfig:"CONVITE_HORAS" default:"72"`
}

// ChaveTokensBytes devolve a chave já decodificada (validada em Load)
func (s SegurancaConfig_t) ChaveTokensBytes() []byte {
	b, _ := base64.StdEncoding.DecodeString(s.ChaveTokens)
	return b
}

func (s SegurancaConfig_t) validar() error {
	b, err := base64.StdEncoding.DecodeString(s.ChaveTokens)
	if err != nil || len(b) != 32 {
		return errors.New("CHAVE_TOKENS precisa ter 32 bytes em base64 (gere com: openssl rand -base64 32)")
	}
	if s.SessaoHoras < 1 || s.ConviteHoras < 1 {
		return errors.New("SESSAO_HORAS e CONVITE_HORAS precisam ser maiores que zero")
	}
	return nil
}
