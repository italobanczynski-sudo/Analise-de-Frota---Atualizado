package usuario

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
)

var (
	ErrNome            = errors.New("informe o nome da pessoa (até 120 letras)")
	ErrEmail           = errors.New("e-mail inválido")
	ErrPapel           = errors.New("perfil inválido: use admin, colaborador ou leitor")
	ErrEmailRepetido   = errors.New("já existe um usuário com esse e-mail")
	ErrNaoEncontrado   = errors.New("usuário não encontrado")
	ErrSenhaFraca      = errors.New("a senha precisa ter pelo menos 10 caracteres, com letras e números")
	ErrSenhaLonga      = errors.New("a senha pode ter no máximo 72 caracteres")
	ErrConviteInvalido = errors.New("este link de convite não vale mais (já foi usado ou venceu); peça um novo a um administrador")
	ErrLogin           = errors.New("e-mail ou senha incorretos")
	ErrInativo         = errors.New("este usuário está desativado")
	ErrUltimoAdmin     = errors.New("a plataforma precisa de pelo menos um administrador ativo")
	ErrCodigoInvalido  = errors.New("código incorreto ou vencido; confira o e-mail ou peça um novo")
	ErrAguarde         = errors.New("aguarde um minuto para pedir outro código")
)

// normalizarCodigo aceita "123 456" ou "123-456"; devolve vazio se não forem 6 dígitos
func normalizarCodigo(c string) string {
	d := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' {
			return -1
		}
		return 'x'
	}, strings.TrimSpace(c))
	if len(d) != 6 || strings.Contains(d, "x") {
		return ""
	}
	return d
}

func normalizarEmail(e string) string { return strings.ToLower(strings.TrimSpace(e)) }

func validarCadastro(nome, email, papel string) error {
	n := []rune(strings.TrimSpace(nome))
	if len(n) == 0 || len(n) > 120 {
		return ErrNome
	}
	if err := validarEmail(email); err != nil {
		return err
	}
	return validarPapel(papel)
}

func validarEmail(email string) error {
	a, err := mail.ParseAddress(email)
	if err != nil || a.Address != email || !strings.Contains(email[strings.LastIndex(email, "@"):], ".") || len(email) > 200 {
		return ErrEmail
	}
	return nil
}

func validarPapel(p string) error {
	switch p {
	case PapelAdmin, PapelColaborador, PapelLeitor:
		return nil
	}
	return ErrPapel
}

// a senha vai para o bcrypt, que só considera os primeiros 72 bytes
func validarSenha(s string) error {
	if len(s) > 72 {
		return ErrSenhaLonga
	}
	letra, numero := false, false
	for _, r := range s {
		letra = letra || unicode.IsLetter(r)
		numero = numero || unicode.IsDigit(r)
	}
	if len([]rune(s)) < 10 || !letra || !numero {
		return ErrSenhaFraca
	}
	return nil
}
