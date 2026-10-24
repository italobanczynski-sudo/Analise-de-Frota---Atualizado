// Package usuario cuida de quem usa a plataforma: cadastro por convite, senha e perfil de acesso.
package usuario

import (
	"context"
	"crypto/hmac"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"analisefrota/common/cripto"
	"analisefrota/environment"
	"analisefrota/modules/usuario/usuario_repo"
)

type Service_I interface {
	Listar(ctx context.Context) ([]Usuario_t, error)
	Buscar(ctx context.Context, id int64) (Usuario_t, error)
	// Criar cadastra a pessoa e já gera o convite para ela criar a senha
	Criar(ctx context.Context, nome, email, papel string) (Usuario_t, Convite_t, error)
	// NovoConvite gera outro link (o anterior deixa de valer); serve também para trocar a senha esquecida
	NovoConvite(ctx context.Context, id int64) (Convite_t, error)
	// AceitarConvite grava a senha escolhida e devolve a pessoa (o link deixa de valer)
	AceitarConvite(ctx context.Context, token, senha string) (Usuario_t, error)
	// Autenticar confere e-mail e senha
	Autenticar(ctx context.Context, email, senha string) (Usuario_t, error)
	// Alterar troca o perfil e ativa/desativa; nunca deixa a plataforma sem administrador ativo
	Alterar(ctx context.Context, id int64, papel string, ativo bool) (Usuario_t, error)
	// SolicitarCodigo gera o código de 6 dígitos (primeiro acesso ou senha esquecida) para quem já foi cadastrado
	SolicitarCodigo(ctx context.Context, email string) (Codigo_t, error)
	// UsarCodigo confere o código, grava a senha escolhida e devolve a pessoa
	UsarCodigo(ctx context.Context, email, codigo, senha string) (Usuario_t, error)
}

type usuario_t struct {
	repo        usuario_repo.Repository_I
	agora       func() time.Time
	validade    time.Duration
	custoBcrypt int
	chave       []byte // chave do HMAC dos códigos
}

func NewUsuario() Service_I {
	return &usuario_t{repo: usuario_repo.NewRepository(), agora: time.Now,
		validade: time.Duration(environment.Cfg.Seguranca.ConviteHoras) * time.Hour, custoBcrypt: 12,
		chave: environment.Cfg.Seguranca.ChaveTokensBytes()}
}

// NewUsuarioCom recebe as dependências (usado nos testes, com bcrypt barato)
func NewUsuarioCom(repo usuario_repo.Repository_I, agora func() time.Time, validade time.Duration, chave []byte) Service_I {
	return &usuario_t{repo: repo, agora: agora, validade: validade, custoBcrypt: bcrypt.MinCost, chave: chave}
}

func (u *usuario_t) SolicitarCodigo(ctx context.Context, email string) (Codigo_t, error) {
	r, err := u.repo.BuscarPorEmail(ctx, normalizarEmail(email))
	if err != nil {
		return Codigo_t{}, traduzir(err)
	}
	if !r.Ativo {
		return Codigo_t{}, ErrInativo
	}
	agora := u.agora()
	ultimo, err := u.repo.UltimoCodigoEm(ctx, r.ID)
	if err != nil {
		return Codigo_t{}, err
	}
	if !ultimo.IsZero() && agora.Sub(ultimo) < CodigoIntervalo {
		return Codigo_t{}, ErrAguarde
	}
	codigo, err := gerarCodigo()
	if err != nil {
		return Codigo_t{}, err
	}
	expira := agora.Add(CodigoValidade)
	if err := u.repo.CriarCodigo(ctx, r.ID, hashCodigo(u.chave, r.ID, codigo), expira, agora); err != nil {
		return Codigo_t{}, err
	}
	return Codigo_t{Codigo: codigo, Usuario: deRow(r), ExpiraEm: expira}, nil
}

func (u *usuario_t) UsarCodigo(ctx context.Context, email, codigo, senha string) (Usuario_t, error) {
	if err := validarSenha(senha); err != nil {
		return Usuario_t{}, err
	}
	codigo = normalizarCodigo(codigo)
	if codigo == "" {
		return Usuario_t{}, ErrCodigoInvalido
	}
	// e-mail desconhecido ou desativado dá a mesma resposta de código errado (não revela quem tem conta)
	r, err := u.repo.BuscarPorEmail(ctx, normalizarEmail(email))
	if err != nil || !r.Ativo {
		if err != nil && !errors.Is(err, usuario_repo.ErrNaoEncontrado) {
			return Usuario_t{}, err
		}
		return Usuario_t{}, ErrCodigoInvalido
	}
	agora := u.agora()
	c, err := u.repo.CodigoAtivo(ctx, r.ID, agora)
	if err != nil {
		if errors.Is(err, usuario_repo.ErrSemCodigo) {
			return Usuario_t{}, ErrCodigoInvalido
		}
		return Usuario_t{}, err
	}
	if c.Tentativas >= CodigoTentativas {
		return Usuario_t{}, ErrCodigoInvalido
	}
	if !hmac.Equal(c.CodigoHash, hashCodigo(u.chave, r.ID, codigo)) {
		_ = u.repo.SomarTentativa(ctx, c.ID)
		return Usuario_t{}, ErrCodigoInvalido
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), u.custoBcrypt)
	if err != nil {
		return Usuario_t{}, err
	}
	if err := u.repo.UsarCodigo(ctx, c.ID, r.ID, string(hash), agora); err != nil {
		if errors.Is(err, usuario_repo.ErrSemCodigo) {
			return Usuario_t{}, ErrCodigoInvalido
		}
		return Usuario_t{}, err
	}
	_ = u.repo.MarcarAcesso(ctx, r.ID, agora)
	return u.Buscar(ctx, r.ID)
}

// hash fixo para gastar o mesmo tempo quando o e-mail não existe (não revela quem tem conta)
var hashFalso, _ = bcrypt.GenerateFromPassword([]byte("senha-que-ninguem-usa-1"), 12)

func (u *usuario_t) Listar(ctx context.Context) ([]Usuario_t, error) {
	rows, err := u.repo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Usuario_t, len(rows))
	for i, r := range rows {
		out[i] = deRow(r)
	}
	return out, nil
}

func (u *usuario_t) Buscar(ctx context.Context, id int64) (Usuario_t, error) {
	r, err := u.repo.BuscarPorID(ctx, id)
	return deRow(r), traduzir(err)
}

func (u *usuario_t) Criar(ctx context.Context, nome, email, papel string) (Usuario_t, Convite_t, error) {
	nome, email = strings.TrimSpace(nome), normalizarEmail(email)
	if err := validarCadastro(nome, email, papel); err != nil {
		return Usuario_t{}, Convite_t{}, err
	}
	id, err := u.repo.Criar(ctx, nome, email, papel)
	if err != nil {
		return Usuario_t{}, Convite_t{}, traduzir(err)
	}
	conv, err := u.NovoConvite(ctx, id)
	if err != nil {
		return Usuario_t{}, Convite_t{}, err
	}
	usr, err := u.Buscar(ctx, id)
	return usr, conv, err
}

func (u *usuario_t) NovoConvite(ctx context.Context, id int64) (Convite_t, error) {
	r, err := u.repo.BuscarPorID(ctx, id)
	if err != nil {
		return Convite_t{}, traduzir(err)
	}
	if !r.Ativo {
		return Convite_t{}, ErrInativo
	}
	token, err := cripto.TokenAleatorio()
	if err != nil {
		return Convite_t{}, err
	}
	agora := u.agora()
	conv := Convite_t{Token: token, ExpiraEm: agora.Add(u.validade)}
	if err := u.repo.CriarConvite(ctx, id, cripto.HashToken(token), conv.ExpiraEm, agora); err != nil {
		return Convite_t{}, err
	}
	return conv, nil
}

func (u *usuario_t) AceitarConvite(ctx context.Context, token, senha string) (Usuario_t, error) {
	if err := validarSenha(senha); err != nil {
		return Usuario_t{}, err
	}
	if strings.TrimSpace(token) == "" {
		return Usuario_t{}, ErrConviteInvalido
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), u.custoBcrypt)
	if err != nil {
		return Usuario_t{}, err
	}
	id, err := u.repo.UsarConvite(ctx, cripto.HashToken(token), string(hash), u.agora())
	if err != nil {
		return Usuario_t{}, traduzir(err)
	}
	return u.Buscar(ctx, id)
}

func (u *usuario_t) Autenticar(ctx context.Context, email, senha string) (Usuario_t, error) {
	r, err := u.repo.BuscarPorEmail(ctx, normalizarEmail(email))
	if err != nil && !errors.Is(err, usuario_repo.ErrNaoEncontrado) {
		return Usuario_t{}, err
	}
	if err != nil || !r.SenhaHash.Valid {
		_ = bcrypt.CompareHashAndPassword(hashFalso, []byte(senha))
		return Usuario_t{}, ErrLogin
	}
	if bcrypt.CompareHashAndPassword([]byte(r.SenhaHash.String), []byte(senha)) != nil {
		return Usuario_t{}, ErrLogin
	}
	if !r.Ativo {
		return Usuario_t{}, ErrInativo
	}
	_ = u.repo.MarcarAcesso(ctx, r.ID, u.agora())
	return deRow(r), nil
}

func (u *usuario_t) Alterar(ctx context.Context, id int64, papel string, ativo bool) (Usuario_t, error) {
	if err := validarPapel(papel); err != nil {
		return Usuario_t{}, err
	}
	atual, err := u.repo.BuscarPorID(ctx, id)
	if err != nil {
		return Usuario_t{}, traduzir(err)
	}
	deixaDeSerAdmin := atual.Papel == PapelAdmin && atual.Ativo && (papel != PapelAdmin || !ativo)
	if deixaDeSerAdmin {
		n, err := u.repo.ContarAdminsAtivos(ctx)
		if err != nil {
			return Usuario_t{}, err
		}
		if n <= 1 {
			return Usuario_t{}, ErrUltimoAdmin
		}
	}
	if err := u.repo.Alterar(ctx, id, papel, ativo); err != nil {
		return Usuario_t{}, traduzir(err)
	}
	return u.Buscar(ctx, id)
}

func deRow(r usuario_repo.UsuarioRow_t) Usuario_t {
	return Usuario_t{ID: r.ID, Nome: r.Nome, Email: r.Email, Papel: r.Papel, Ativo: r.Ativo, TemSenha: r.SenhaHash.Valid,
		CriadoEm: r.CriadoEm, UltimoAcesso: r.UltimoAcesso}
}

func traduzir(err error) error {
	switch {
	case errors.Is(err, usuario_repo.ErrNaoEncontrado):
		return ErrNaoEncontrado
	case errors.Is(err, usuario_repo.ErrRepetido):
		return ErrEmailRepetido
	case errors.Is(err, usuario_repo.ErrConviteInvalido):
		return ErrConviteInvalido
	}
	return err
}
