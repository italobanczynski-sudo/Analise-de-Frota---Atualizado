package sessao

import "errors"

var ErrSessaoInvalida = errors.New("sessão vencida ou inválida; entre de novo")

// tokens gerados por cripto.TokenAleatorio têm 43 caracteres; qualquer outra coisa nem vai ao banco
func formatoValido(token string) bool { return len(token) == 43 }
