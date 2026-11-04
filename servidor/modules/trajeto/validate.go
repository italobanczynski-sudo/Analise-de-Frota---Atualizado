package trajeto

import (
	"errors"
	"regexp"
	"time"

	"analisefrota/common/utils"
)

var (
	ErrPlaca     = errors.New("placa inválida: use só letras e números, como ABC1D23")
	ErrDia       = errors.New("dia inválido: use AAAA-MM-DD")
	ErrDiaFuturo = errors.New("o dia do trajeto não pode ser no futuro")
	ErrDiaAntigo = errors.New("o trajeto só pode ser pedido até 93 dias para trás")
	ErrHoras     = errors.New("horas inválidas: de 0 a 24, com o início antes do fim")
	placaValida  = regexp.MustCompile(`^[A-Z0-9]{5,8}$`)
)

// validar confere o pedido e devolve a placa limpa (sem hífen, maiúscula)
func validar(placa, dia string, de, ate int, agora time.Time) (string, error) {
	p := utils.ChavePlaca(placa)
	if !placaValida.MatchString(p) {
		return "", ErrPlaca
	}
	if _, err := time.Parse(utils.LayoutDia, dia); err != nil {
		return "", ErrDia
	}
	hoje := utils.DiaBR(agora)
	if dia > hoje {
		return "", ErrDiaFuturo
	}
	if dia < utils.SomarDias(hoje, -DiasMax) {
		return "", ErrDiaAntigo
	}
	if de < 0 || ate > 24 || de >= ate {
		return "", ErrHoras
	}
	return p, nil
}
