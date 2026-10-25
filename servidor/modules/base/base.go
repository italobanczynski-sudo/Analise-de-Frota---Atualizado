// Package base guarda os documentos do painel (index.html) que ficavam na base compartilhada do Artifact.
package base

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"analisefrota/modules/base/base_repo"
)

type Service_I interface {
	Listar(ctx context.Context, colecao string) ([]Documento_t, error)
	// Gravar cria ou substitui o documento (como o set do painel)
	Gravar(ctx context.Context, colecao, id string, dados []byte, usuarioID int64) error
	// Criar grava com um id novo (como o add do painel) e devolve o id
	Criar(ctx context.Context, colecao string, dados []byte, usuarioID int64) (string, error)
	Apagar(ctx context.Context, colecao, id string) error
}

type base_t struct{ repo base_repo.Repository_I }

func NewBase() Service_I { return &base_t{repo: base_repo.NewRepository()} }

// NewBaseCom recebe o repositório (usado nos testes)
func NewBaseCom(repo base_repo.Repository_I) Service_I { return &base_t{repo: repo} }

func (b *base_t) Listar(ctx context.Context, colecao string) ([]Documento_t, error) {
	if err := validarColecao(colecao); err != nil {
		return nil, err
	}
	rows, err := b.repo.Listar(ctx, colecao)
	if err != nil {
		return nil, err
	}
	out := make([]Documento_t, len(rows))
	for i, r := range rows {
		out[i] = Documento_t{ID: r.ID, Dados: r.Dados, AtualizadoEm: r.AtualizadoEm}
	}
	return out, nil
}

func (b *base_t) Gravar(ctx context.Context, colecao, id string, dados []byte, usuarioID int64) error {
	if err := validarDocumento(colecao, id, dados); err != nil {
		return err
	}
	return b.repo.Gravar(ctx, colecao, id, dados, usuarioID)
}

func (b *base_t) Criar(ctx context.Context, colecao string, dados []byte, usuarioID int64) (string, error) {
	r := make([]byte, 10)
	if _, err := rand.Read(r); err != nil {
		return "", err
	}
	id := hex.EncodeToString(r)
	return id, b.Gravar(ctx, colecao, id, dados, usuarioID)
}

func (b *base_t) Apagar(ctx context.Context, colecao, id string) error {
	if err := validarColecao(colecao); err != nil {
		return err
	}
	if !idValido.MatchString(id) {
		return ErrID
	}
	err := b.repo.Apagar(ctx, colecao, id)
	if errors.Is(err, base_repo.ErrNaoEncontrado) {
		return nil // apagar o que já não existe não é erro (o painel apaga partes que podem não existir)
	}
	return err
}
