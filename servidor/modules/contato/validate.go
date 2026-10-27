package contato

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrMotorista      = errors.New("informe o nome do motorista como aparece na API Gobrax (até 120 letras)")
	ErrTelefone       = errors.New("telefone inválido: use DDD + número, por exemplo (11) 91234-5678")
	ErrSemAutorizacao = errors.New("confirme que o motorista autorizou receber mensagens no WhatsApp")
	ErrNaoEncontrado  = errors.New("motorista sem WhatsApp cadastrado")
	naoDigito         = regexp.MustCompile(`\D`)
	// 55 + DDD (11 a 99) + celular (9 e mais 8 dígitos) ou fixo (2 a 5 e mais 7 dígitos)
	telefoneBR = regexp.MustCompile(`^55[1-9][1-9](9[0-9]{8}|[2-5][0-9]{7})$`)
)

func validarMotorista(nome string) error {
	if nome == "" || len([]rune(nome)) > 120 {
		return ErrMotorista
	}
	return nil
}

// NormalizarTelefone devolve o número no formato +55DDNNNNNNNNN; aceita com ou sem +55, espaços, traços e parênteses
func NormalizarTelefone(v string) (string, error) {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "+") && !strings.HasPrefix(naoDigito.ReplaceAllString(v, ""), "55") {
		return "", ErrTelefone // só números do Brasil
	}
	d := naoDigito.ReplaceAllString(v, "")
	if strings.HasPrefix(d, "0") && (len(d) == 11 || len(d) == 12) {
		d = d[1:] // 0 de longa distância: 011...
	}
	if len(d) == 10 || len(d) == 11 {
		d = "55" + d
	}
	// celular no formato antigo, sem o 9 na frente (DDD + 8 dígitos começando com 6 a 9): acrescenta o 9
	if len(d) == 12 && d[4] >= '6' {
		d = d[:4] + "9" + d[4:]
	}
	if !telefoneBR.MatchString(d) {
		return "", ErrTelefone
	}
	return "+" + d, nil
}
