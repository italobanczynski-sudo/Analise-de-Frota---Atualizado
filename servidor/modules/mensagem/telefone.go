package mensagem

import (
	"context"
	"errors"
	"sort"
	"strings"
	"unicode"

	"analisefrota/common/utils"
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

// DiasCondutores: quantos dias para trás contam para a lista de condutores do cliente
const DiasCondutores = 90

// Condutor_t é um motorista do cliente com o telefone (vazio = sem cadastro) e quando rodou por último
type Condutor_t struct {
	Motorista string `json:"motorista"`
	Telefone  string `json:"telefone"`
	UltimoDia string `json:"ultimo_dia"` // vazio = não rodou nos últimos DiasCondutores dias (só tem telefone)
	Dias      int    `json:"dias"`       // dias com registro nos últimos DiasCondutores dias
}

// Condutores lista os motoristas que rodaram nos últimos 90 dias e os que já têm telefone, em ordem alfabética
func (m *mensagem_t) Condutores(ctx context.Context, clienteID int64) ([]Condutor_t, error) {
	hoje := utils.DiaBR(m.d.Agora())
	regs, err := m.d.Diario.Listar(ctx, clienteID, utils.SomarDias(hoje, -DiasCondutores), hoje)
	if err != nil {
		return nil, err
	}
	tels, err := m.d.Telefones.Linhas(ctx, clienteID)
	if err != nil {
		return nil, err
	}
	por := map[string]*Condutor_t{}
	diasVistos := map[string]map[string]bool{}
	for _, r := range regs {
		k := ChaveMotorista(r.Motorista)
		if k == "" || !r.Rodou {
			continue
		}
		c := por[k]
		if c == nil {
			c = &Condutor_t{}
			por[k], diasVistos[k] = c, map[string]bool{}
		}
		if r.Dia >= c.UltimoDia { // o nome como veio no dia mais recente
			c.Motorista, c.UltimoDia = strings.Join(strings.Fields(r.Motorista), " "), r.Dia
		}
		if !diasVistos[k][r.Dia] {
			diasVistos[k][r.Dia] = true
			c.Dias++
		}
	}
	for _, t := range tels {
		c := por[t.Chave]
		if c == nil {
			c = &Condutor_t{Motorista: t.Motorista}
			por[t.Chave] = c
		}
		c.Telefone = t.Telefone
	}
	out := make([]Condutor_t, 0, len(por))
	for _, c := range por {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return ChaveMotorista(out[i].Motorista) < ChaveMotorista(out[j].Motorista) })
	return out, nil
}
