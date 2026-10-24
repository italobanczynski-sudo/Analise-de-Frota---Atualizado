package coleta

import "errors"

var (
	ErrEmAndamento = errors.New("a busca deste cliente já está em andamento")
	ErrDiasColeta  = errors.New("a primeira busca precisa ser de 1 a 93 dias")
)

func validarDias(dias int) error {
	if dias < 1 || dias > 93 {
		return ErrDiasColeta
	}
	return nil
}
