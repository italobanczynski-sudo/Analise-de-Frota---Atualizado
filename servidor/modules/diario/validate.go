package diario

import (
	"errors"
	"time"

	"analisefrota/common/utils"
)

var (
	ErrDiaInvalido      = errors.New("data inválida: use AAAA-MM-DD")
	ErrPeriodoInvertido = errors.New("o início do período é depois do fim")
	ErrPeriodoLongo     = errors.New("consulte no máximo 400 dias por vez")
)

func validarPeriodo(ini, fim string) error {
	a, err1 := time.Parse(utils.LayoutDia, ini)
	b, err2 := time.Parse(utils.LayoutDia, fim)
	if err1 != nil || err2 != nil {
		return ErrDiaInvalido
	}
	if a.After(b) {
		return ErrPeriodoInvertido
	}
	if b.Sub(a) > 400*24*time.Hour {
		return ErrPeriodoLongo
	}
	return nil
}
