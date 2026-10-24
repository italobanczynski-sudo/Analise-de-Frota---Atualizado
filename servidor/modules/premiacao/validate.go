package premiacao

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
)

var (
	ErrMesInvalido = errors.New("mês inválido: use AAAA-MM")
	mesAPI         = regexp.MustCompile(`^(\d{2})-(\d{4})$`)
	mesPainel      = regexp.MustCompile(`^\d{4}-\d{2}$`)
)

// mesDaAPI converte "09-2026" em "2026-09"; vazio se o formato for outro
func mesDaAPI(s string) string {
	m := mesAPI.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	return m[2] + "-" + m[1]
}

func validarMes(mes string) error {
	if mes != "" && !mesPainel.MatchString(mes) {
		return ErrMesInvalido
	}
	return nil
}

// comErro: a API marca o mês com problema no campo Err (qualquer valor que não seja vazio, null, false ou 0)
func comErro(err json.RawMessage) bool {
	e := bytes.TrimSpace(err)
	if len(e) == 0 {
		return false
	}
	switch string(e) {
	case "null", "false", "0", `""`:
		return false
	}
	return true
}
