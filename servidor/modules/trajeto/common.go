package trajeto

import (
	"sort"
	"time"

	"analisefrota/common/utils"
	"analisefrota/modules/trajeto/trajeto_repo"
)

// janelas devolve o início de cada janela de 5 min entre as horas de e ate do dia de Brasília (em UTC);
// janelas que ainda não começaram ficam de fora
func janelas(dia string, de, ate int, agora time.Time) ([]time.Time, error) {
	ini, _, err := utils.JanelaDiaBR(dia)
	if err != nil {
		return nil, ErrDia
	}
	out := []time.Time{}
	fim := ini.Add(time.Duration(ate) * time.Hour)
	for t := ini.Add(time.Duration(de) * time.Hour); t.Before(fim); t = t.Add(Passo) {
		if t.After(agora) {
			break
		}
		out = append(out, t)
	}
	return out, nil
}

// definitiva: a janela já tinha acabado havia margemFinal quando foi buscada, então o que veio não muda mais
func definitiva(row trajeto_repo.PosicaoRow_t) bool {
	fimJanela := row.JanelaIni.Add(Passo - time.Second)
	return fimJanela.Before(row.BuscadoEm.Add(-margemFinal))
}

// pontosDe junta as janelas com posição, sem repetir a mesma coleta (veículo parado repete a última posição
// em várias janelas), em ordem de data
func pontosDe(rows []trajeto_repo.PosicaoRow_t) []Ponto_t {
	vistos := map[string]bool{}
	out := []Ponto_t{}
	for _, r := range rows {
		if r.Vazio || r.Lat == nil || r.Lon == nil || r.Data == "" || vistos[r.Data] {
			continue
		}
		vistos[r.Data] = true
		out = append(out, Ponto_t{Data: r.Data, Hora: horaBR(r.Data), Lat: *r.Lat, Lon: *r.Lon})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Data < out[j].Data })
	return out
}

// horaBR lê a data da API (UTC) e devolve a hora de Brasília; formato inesperado = vazio
func horaBR(data string) string {
	t, err := time.ParseInLocation(utils.LayoutAPI, data, time.UTC)
	if err != nil {
		return ""
	}
	return t.Add(-3 * time.Hour).Format("15:04:05")
}
