package base

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"slices"
)

var (
	ErrColecao       = errors.New("coleção desconhecida na base do painel")
	ErrID            = errors.New("id de documento inválido")
	ErrDados         = errors.New("o documento precisa ser um objeto JSON")
	ErrGrande        = errors.New("o documento passa de 1 MB; divida a planilha em partes menores")
	ErrNaoEncontrado = errors.New("documento não encontrado")
	idValido         = regexp.MustCompile(`^[A-Za-z0-9._~-]{1,200}$`)
)

func validarColecao(c string) error {
	if !slices.Contains(Colecoes, c) {
		return ErrColecao
	}
	return nil
}

func validarDocumento(colecao, id string, dados []byte) error {
	if err := validarColecao(colecao); err != nil {
		return err
	}
	if !idValido.MatchString(id) {
		return ErrID
	}
	if len(dados) > TamanhoMaximo {
		return ErrGrande
	}
	d := bytes.TrimSpace(dados)
	if len(d) == 0 || d[0] != '{' || !json.Valid(d) {
		return ErrDados
	}
	return nil
}
