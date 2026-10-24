// Package cliente cuida dos clientes acompanhados e do token da API de cada um.
// O token é cifrado antes de ir para o banco e só é decifrado para chamar a API.
package cliente

import (
	"context"
	"errors"
	"strings"

	"analisefrota/common/cripto"
	"analisefrota/common/utils"
	"analisefrota/environment"
	"analisefrota/modules/cliente/cliente_repo"
)

type Service_I interface {
	Listar(ctx context.Context) ([]Cliente_t, error)
	Buscar(ctx context.Context, id int64) (Cliente_t, error)
	BuscarPorNome(ctx context.Context, nome string) (Cliente_t, error)
	Criar(ctx context.Context, nome, hubspotID string) (Cliente_t, error)
	SalvarToken(ctx context.Context, id int64, token string) error
	// Token devolve o token decifrado, só para chamar a API Gobrax; nunca vai para uma resposta
	Token(ctx context.Context, id int64) (string, error)
}

type cliente_t struct {
	repo  cliente_repo.Repository_I
	chave []byte
}

func NewCliente() Service_I {
	return &cliente_t{repo: cliente_repo.NewRepository(), chave: environment.Cfg.Seguranca.ChaveTokensBytes()}
}

// NewClienteCom recebe o repositório e a chave (usado nos testes)
func NewClienteCom(repo cliente_repo.Repository_I, chave []byte) Service_I {
	return &cliente_t{repo: repo, chave: chave}
}

func (c *cliente_t) Listar(ctx context.Context) ([]Cliente_t, error) {
	rows, err := c.repo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Cliente_t, len(rows))
	for i, r := range rows {
		out[i] = deRow(r)
	}
	return out, nil
}

func (c *cliente_t) Buscar(ctx context.Context, id int64) (Cliente_t, error) {
	r, err := c.repo.BuscarPorID(ctx, id)
	return deRow(r), traduzir(err)
}

func (c *cliente_t) BuscarPorNome(ctx context.Context, nome string) (Cliente_t, error) {
	r, err := c.repo.BuscarPorNome(ctx, strings.TrimSpace(nome))
	return deRow(r), traduzir(err)
}

func (c *cliente_t) Criar(ctx context.Context, nome, hubspotID string) (Cliente_t, error) {
	nome, hubspotID = strings.TrimSpace(nome), strings.TrimSpace(hubspotID)
	if err := validarNome(nome); err != nil {
		return Cliente_t{}, err
	}
	if err := validarHubspot(hubspotID); err != nil {
		return Cliente_t{}, err
	}
	id, err := c.repo.Criar(ctx, nome, hubspotID)
	if err != nil {
		return Cliente_t{}, traduzir(err)
	}
	return c.Buscar(ctx, id)
}

func (c *cliente_t) SalvarToken(ctx context.Context, id int64, token string) error {
	token = utils.LimparToken(token)
	if err := validarToken(token); err != nil {
		return err
	}
	cifrado, err := cripto.Cifrar(c.chave, token)
	if err != nil {
		return err
	}
	return traduzir(c.repo.SalvarTokenCifrado(ctx, id, cifrado))
}

func (c *cliente_t) Token(ctx context.Context, id int64) (string, error) {
	cifrado, err := c.repo.TokenCifrado(ctx, id)
	if err != nil {
		return "", traduzir(err)
	}
	if len(cifrado) == 0 {
		return "", ErrSemToken
	}
	return cripto.Decifrar(c.chave, cifrado)
}

func deRow(r cliente_repo.ClienteRow_t) Cliente_t {
	return Cliente_t{ID: r.ID, Nome: r.Nome, HubspotID: r.HubspotID, TemToken: r.TemToken, CriadoEm: r.CriadoEm}
}

func traduzir(err error) error {
	switch {
	case errors.Is(err, cliente_repo.ErrNaoEncontrado):
		return ErrNaoEncontrado
	case errors.Is(err, cliente_repo.ErrRepetido):
		return ErrNomeRepetido
	}
	return err
}
