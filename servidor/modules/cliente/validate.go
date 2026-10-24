package cliente

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrNomeVazio     = errors.New("informe o nome do cliente")
	ErrNomeLongo     = errors.New("o nome do cliente passa de 120 caracteres")
	ErrHubspotID     = errors.New("o ID do HubSpot tem só números (o número no fim do link da empresa)")
	ErrTokenVazio    = errors.New("informe o token da API Gobrax")
	ErrTokenEstranho = errors.New("o token tem caracteres que não são de um token; copie de novo, inteiro")
	ErrNaoEncontrado = errors.New("cliente não encontrado")
	ErrSemToken      = errors.New("este cliente ainda não tem token da API salvo")
	ErrNomeRepetido  = errors.New("já existe um cliente com esse nome")
	tokenValido      = regexp.MustCompile(`^[A-Za-z0-9._\-+/=]+$`)
	somenteNumeros   = regexp.MustCompile(`^[0-9]*$`)
)

func validarNome(nome string) error {
	nome = strings.TrimSpace(nome)
	if nome == "" {
		return ErrNomeVazio
	}
	if len([]rune(nome)) > 120 {
		return ErrNomeLongo
	}
	return nil
}

func validarHubspot(id string) error {
	if !somenteNumeros.MatchString(strings.TrimSpace(id)) {
		return ErrHubspotID
	}
	return nil
}

func validarToken(token string) error {
	if token == "" {
		return ErrTokenVazio
	}
	if !tokenValido.MatchString(token) {
		return ErrTokenEstranho
	}
	return nil
}
