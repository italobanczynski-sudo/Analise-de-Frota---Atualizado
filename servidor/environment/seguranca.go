package environment

import (
	"encoding/base64"
	"errors"
)

type SegurancaConfig_t struct {
	// Chave AES-256 (32 bytes em base64) que cifra os tokens dos clientes no banco
	ChaveTokens string `envconfig:"CHAVE_TOKENS" required:"true"`
	// Chave enviada no cabeçalho X-Chave-Admin até a fase de login (fase 3)
	ChaveAdmin string `envconfig:"CHAVE_ADMIN" required:"true"`
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
	if len(s.ChaveAdmin) < 24 {
		return errors.New("CHAVE_ADMIN precisa ter pelo menos 24 caracteres")
	}
	return nil
}
