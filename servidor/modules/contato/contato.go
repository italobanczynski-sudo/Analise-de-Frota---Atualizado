// Package contato guarda o WhatsApp de cada motorista e a autorização para receber mensagens.
// O telefone é cifrado antes de ir para o banco e só é decifrado na hora do envio; a tela vê só os 4 últimos dígitos.
package contato

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/common/cripto"
	"analisefrota/environment"
	"analisefrota/modules/contato/contato_repo"
)

type Service_I interface {
	Listar(ctx context.Context, clienteID int64) ([]Contato_t, error)
	// Salvar cadastra ou troca o telefone; exige a confirmação de que o motorista autorizou
	Salvar(ctx context.Context, clienteID int64, motorista, telefone string, autorizado bool, usuarioID int64) (Contato_t, error)
	Ativar(ctx context.Context, clienteID int64, motorista string, ativo bool) error
	Remover(ctx context.Context, clienteID int64, motorista string) error
	// Destinos devolve os motoristas ativos com o telefone decifrado, só para enviar
	Destinos(ctx context.Context, clienteID int64) ([]Destino_t, error)
}

type contato_t struct {
	repo  contato_repo.Repository_I
	chave []byte
	agora func() time.Time
}

func NewContato() Service_I {
	return &contato_t{repo: contato_repo.NewRepository(), chave: environment.Cfg.Seguranca.ChaveTokensBytes(), agora: time.Now}
}

// NewContatoCom recebe o repositório, a chave e o relógio (usado nos testes)
func NewContatoCom(repo contato_repo.Repository_I, chave []byte, agora func() time.Time) Service_I {
	return &contato_t{repo: repo, chave: chave, agora: agora}
}

func (c *contato_t) Listar(ctx context.Context, clienteID int64) ([]Contato_t, error) {
	rows, err := c.repo.Listar(ctx, clienteID)
	if err != nil {
		return nil, err
	}
	out := make([]Contato_t, len(rows))
	for i, r := range rows {
		out[i] = deRow(r)
	}
	return out, nil
}

func (c *contato_t) Salvar(ctx context.Context, clienteID int64, motorista, telefone string, autorizado bool, usuarioID int64) (Contato_t, error) {
	motorista = strings.TrimSpace(motorista)
	if err := validarMotorista(motorista); err != nil {
		return Contato_t{}, err
	}
	tel, err := NormalizarTelefone(telefone)
	if err != nil {
		return Contato_t{}, err
	}
	if !autorizado {
		return Contato_t{}, ErrSemAutorizacao
	}
	cifrado, err := cripto.Cifrar(c.chave, tel)
	if err != nil {
		return Contato_t{}, err
	}
	row := contato_repo.ContatoRow_t{ClienteID: clienteID, Motorista: motorista, TelefoneCifrado: cifrado,
		TelefoneFinal: tel[len(tel)-4:], Ativo: true, AutorizadoEm: c.agora()}
	if usuarioID > 0 {
		row.AutorizadoPor = &usuarioID
	}
	if err := c.repo.Salvar(ctx, row); err != nil {
		return Contato_t{}, err
	}
	return deRow(row), nil
}

func (c *contato_t) Ativar(ctx context.Context, clienteID int64, motorista string, ativo bool) error {
	return traduzir(c.repo.Ativar(ctx, clienteID, strings.TrimSpace(motorista), ativo))
}

func (c *contato_t) Remover(ctx context.Context, clienteID int64, motorista string) error {
	return traduzir(c.repo.Remover(ctx, clienteID, strings.TrimSpace(motorista)))
}

func (c *contato_t) Destinos(ctx context.Context, clienteID int64) ([]Destino_t, error) {
	rows, err := c.repo.ListarAtivos(ctx, clienteID)
	if err != nil {
		return nil, err
	}
	out := make([]Destino_t, 0, len(rows))
	for _, r := range rows {
		tel, err := cripto.Decifrar(c.chave, r.TelefoneCifrado)
		if err != nil { // chave trocada: esse motorista fica sem mensagem até o telefone ser salvo de novo
			logrus.Warnf("contato: telefone de um motorista do cliente %d não decifra (CHAVE_TOKENS trocada?); salve de novo", clienteID)
			continue
		}
		out = append(out, Destino_t{Motorista: r.Motorista, Telefone: tel})
	}
	return out, nil
}

func deRow(r contato_repo.ContatoRow_t) Contato_t {
	return Contato_t{Motorista: r.Motorista, TelefoneFinal: r.TelefoneFinal, Ativo: r.Ativo, AutorizadoEm: r.AutorizadoEm}
}

func traduzir(err error) error {
	if errors.Is(err, contato_repo.ErrNaoEncontrado) {
		return ErrNaoEncontrado
	}
	return err
}
