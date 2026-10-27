package utils

import "time"

const (
	LayoutDia = "2006-01-02"
	// Formato das datas que a API Gobrax recebe (sempre em UTC)
	LayoutAPI = "2006-01-02 15:04:05"
)

// Brasília é UTC−3 o ano todo (sem horário de verão desde 2019)
var fusoBR = time.FixedZone("BRT", -3*3600)

// FusoBR devolve o fuso de Brasília (usado pelo agendador das tarefas com hora marcada)
func FusoBR() *time.Location { return fusoBR }

// DiaBR devolve o dia de Brasília (AAAA-MM-DD) do instante t
func DiaBR(t time.Time) string { return t.In(fusoBR).Format(LayoutDia) }

// JanelaDiaBR devolve o início e o fim do dia de Brasília em UTC:
// "2026-10-23" vira 2026-10-23 03:00:00 até 2026-10-24 02:59:59
func JanelaDiaBR(dia string) (time.Time, time.Time, error) {
	d, err := time.ParseInLocation(LayoutDia, dia, fusoBR)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return d.UTC(), d.Add(24*time.Hour - time.Second).UTC(), nil
}

// FormatoAPI escreve o instante como a API espera: UTC, "AAAA-MM-DD HH:MM:SS"
func FormatoAPI(t time.Time) string { return t.UTC().Format(LayoutAPI) }

// SomarDias soma n dias a um dia AAAA-MM-DD
func SomarDias(dia string, n int) string {
	d, err := time.Parse(LayoutDia, dia)
	if err != nil {
		return dia
	}
	return d.AddDate(0, 0, n).Format(LayoutDia)
}

// DiasEntre lista os dias de ini a fim, inclusive
func DiasEntre(ini, fim string) []string {
	var dias []string
	for d := ini; d <= fim; d = SomarDias(d, 1) {
		dias = append(dias, d)
	}
	return dias
}

// MesAPI converte AAAA-MM-DD para o MM-AAAA da premiação
func MesAPI(dia string) string {
	if len(dia) < 7 {
		return ""
	}
	return dia[5:7] + "-" + dia[0:4]
}

// InstanteAPI lê as datas ISO da API ("2026-10-23T03:00:00+0000" ou RFC 3339)
func InstanteAPI(s string) (time.Time, bool) {
	for _, l := range []string{"2006-01-02T15:04:05-0700", time.RFC3339, "2006-01-02 15:04:05-0700"} {
		if t, err := time.Parse(l, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
