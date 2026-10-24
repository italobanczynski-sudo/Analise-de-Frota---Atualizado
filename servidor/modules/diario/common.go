package diario

import (
	"encoding/json"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/common/utils"
)

// motoristaPrincipal: com mais de um motorista no dia, fica quem dirigiu por mais tempo
// (somar a linha para todos contaria os km e as horas em dobro). Empate: o primeiro da lista.
func motoristaPrincipal(p *gobraxapi.PerformanceRec_t) string {
	if p == nil {
		return ""
	}
	ordem := []string{}
	tempo := map[string]float64{}
	for _, d := range p.Drivers {
		nome := d.DriverName
		if nome == "" {
			continue
		}
		dur := 1.0
		ini, ok1 := utils.InstanteAPI(d.StartDate)
		fim, ok2 := utils.InstanteAPI(d.EndDate)
		if ok1 && ok2 && fim.After(ini) {
			dur = float64(fim.Sub(ini).Milliseconds())
		}
		if _, ja := tempo[nome]; !ja {
			ordem = append(ordem, nome)
		}
		tempo[nome] += dur
	}
	melhor := ""
	for _, n := range ordem {
		if melhor == "" || tempo[n] > tempo[melhor] {
			melhor = n
		}
	}
	return melhor
}

// percentual de um indicador; a API escreve "pedalPressureOnHig" (sem o h final) em parte das respostas
func percentual(ind map[string]gobraxapi.Indicador_t, nome string) *float64 {
	x, ok := ind[nome]
	if !ok && nome == "pedalPressureOnHigh" {
		x, ok = ind["pedalPressureOnHig"]
	}
	if !ok {
		return nil
	}
	return arred(x.Percentage.Valor)
}

// nota geral: número ou {"score": n}; só vale entre 0 e 100 (hoje a performance não traz, fica nil)
func notaGeral(p *gobraxapi.PerformanceRec_t) *float64 {
	if p == nil {
		return nil
	}
	bruto := p.Score
	if len(bruto) == 0 {
		if x, ok := p.Indicators["score"]; ok {
			return faixaNota(x.Score.Valor)
		}
		return nil
	}
	var obj struct {
		Score gobraxapi.Numero_t `json:"score"`
	}
	if json.Unmarshal(bruto, &obj) == nil && obj.Score.Valor != nil {
		return faixaNota(obj.Score.Valor)
	}
	var n gobraxapi.Numero_t
	_ = json.Unmarshal(bruto, &n)
	return faixaNota(n.Valor)
}

func faixaNota(v *float64) *float64 {
	if v == nil || *v < 0 || *v > 100 {
		return nil
	}
	return arred(v)
}

func arred(v *float64) *float64 {
	if v == nil {
		return nil
	}
	x := utils.Arred2(*v)
	return &x
}

func valorOu0(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
