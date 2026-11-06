package mensagem

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"analisefrota/modules/mensagem/mensagem_repo"
)

var ErrTelefone = errors.New("telefone inválido: use DDD + número (10 ou 11 dígitos), com ou sem o 55 do Brasil")

// NormalizarTelefone deixa só os dígitos e põe o 55 do Brasil quando vier só DDD + número.
// Vazio continua vazio (apaga o telefone). Aceita 10–11 dígitos (Brasil sem o 55) ou 12–13 começando com 55.
func NormalizarTelefone(t string) (string, error) {
	var b strings.Builder
	for _, r := range t {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	n := strings.TrimLeft(b.String(), "0") // 0 de discagem interurbana (0 11 9…)
	switch {
	case n == "" && strings.TrimSpace(t) == "":
		return "", nil
	case len(n) == 10 || len(n) == 11:
		return "55" + n, nil
	case (len(n) == 12 || len(n) == 13) && strings.HasPrefix(n, "55"):
		return n, nil
	}
	return "", ErrTelefone
}

// SalvarTelefone grava o telefone do motorista; telefone vazio apaga. Devolve o número como ficou gravado.
func (m *mensagem_t) SalvarTelefone(ctx context.Context, clienteID int64, motorista, telefone string, usuarioID int64) (string, error) {
	n, err := NormalizarTelefone(telefone)
	if err != nil {
		return "", err
	}
	chave := ChaveMotorista(motorista)
	if chave == "" {
		return "", ErrMotoristaSemDia
	}
	if n == "" {
		return "", m.d.Telefones.Apagar(ctx, clienteID, chave)
	}
	var por *int64
	if usuarioID > 0 {
		por = &usuarioID
	}
	return n, m.d.Telefones.Salvar(ctx, mensagem_repo.TelefoneRow_t{ClienteID: clienteID, Chave: chave, Motorista: strings.Join(strings.Fields(motorista), " "), Telefone: n, AtualizadoPor: por})
}
