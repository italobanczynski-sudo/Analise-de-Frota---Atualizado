package mensagem

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"analisefrota/common/utils"
	"analisefrota/modules/diario"
)

// TextoTemplate é o modelo da mensagem; {{n}} são trocados pelas variáveis do motorista (Texto)
const TextoTemplate = `Bom dia, {{1}}! Aqui está a sua análise de {{2}}{{3}}:

🛣️ Km rodados: {{4}}
⛽ Média: {{5}}
🟢 Início de Faixa Verde: {{6}}
🚛 Aproveitamento de Embalo: {{7}}
⏸️ Motor Ligado Parado: {{8}}
🔺 Acelerando acima do Verde: {{9}}
🚨 Excesso de Velocidade: {{10}}
🎯 Piloto Automático: {{11}}
🦶 Pedal Crítico: {{12}}

⭐ Nota do mês: {{13}}

📈 {{14}}

💡 Para hoje: {{15}}

Boa viagem!`

// quantas variáveis o modelo tem
const totalVariaveis = 15

// DiasComparacao: o dia é comparado com o dia anterior em que o motorista rodou, procurado até este tanto de dias para trás
const DiasComparacao = 7

// piora mínima, em pontos percentuais, para um indicador virar a dica do dia
const pioraMinima = 2.0

// indicadores da análise individual: os nomes da plataforma Gobrax. Ao juntar veículos do mesmo dia, cada percentual
// é ponderado pela própria base, como a plataforma (docs/API.md): ignição ligada no motor ligado parado, pé no
// acelerador no pedal crítico e tempo em movimento nos demais. A média de consumo (km/l) é soma dos km ÷ soma dos litros.
type indicador_t struct {
	chave    string // nome da coluna no diário (e no painel)
	variavel string
	rotulo   string
	unidade  string  // "%" ou " km/l"
	casas    int     // casas decimais no texto
	estavel  float64 // diferença que ainda conta como estável
	melhor   int     // 1 = quanto maior melhor; -1 = quanto menor melhor
	valor    func(r diario.Registro_t) *float64
	peso     func(r diario.Registro_t) *float64
	razao    bool
	dica     string
}

var kml = indicador_t{chave: "kml", variavel: "5", rotulo: "Média de consumo", unidade: " km/l", casas: 2, estavel: 0.05, melhor: 1, razao: true}

// percentuais, na ordem do template; só eles entram na dica do dia
var indicadores = []indicador_t{
	{"inicio_faixa_verde", "6", "Início de Faixa Verde", "%", 1, 1, 1, func(r diario.Registro_t) *float64 { return r.InicioFaixaVerde }, horas, false,
		"mantenha a rotação no início da faixa verde: é onde o motor gasta menos diesel."},
	{"embalo", "7", "Aproveitamento de Embalo", "%", 1, 1, 1, func(r diario.Registro_t) *float64 { return r.Embalo }, horas, false,
		"olhe a via com antecedência e tire o pé antes de descidas, curvas e paradas para rodar no embalo."},
	{"motor_ocioso", "8", "Motor Ligado Parado", "%", 1, 1, -1, func(r diario.Registro_t) *float64 { return r.MotorOcioso }, horasMotor, false,
		"desligue o motor em esperas, filas, carga e descarga. Cada hora parada com o motor ligado queima cerca de 5 litros."},
	{"excesso_rotacao", "9", "Acelerando acima do Verde", "%", 1, 1, -1, func(r diario.Registro_t) *float64 { return r.ExcessoRotacao }, horas, false,
		"troque de marcha antes de passar da faixa verde e evite reduzir em rotação alta."},
	{"tempo_excesso_velocidade", "10", "Excesso de Velocidade", "%", 1, 1, -1, func(r diario.Registro_t) *float64 { return r.TempoExcessoVelocidade }, horas, false,
		"respeite o limite de velocidade: andar acima dele aumenta o consumo e o risco."},
	{"piloto_automatico", "11", "Piloto Automático", "%", 1, 1, 1, func(r diario.Registro_t) *float64 { return r.PilotoAutomatico }, horas, false,
		"use o piloto automático nos trechos planos de rodovia para manter a velocidade constante."},
	{"pressao_acelerador", "12", "Pedal Crítico", "%", 1, 1, -1, func(r diario.Registro_t) *float64 { return r.PressaoAcelerador }, horasPedal, false,
		"acelere de forma progressiva: pedal no fundo gasta muito e ganha pouca velocidade."},
}

func horas(r diario.Registro_t) *float64 { return r.Horas }
func km(r diario.Registro_t) *float64    { return r.Km }

// bases próprias; dia gravado antes de a base existir usa as horas em movimento (como o painel)
func horasMotor(r diario.Registro_t) *float64 {
	if r.HorasMotor != nil {
		return r.HorasMotor
	}
	return r.Horas
}

func horasPedal(r diario.Registro_t) *float64 {
	if r.HorasPedal != nil {
		return r.HorasPedal
	}
	return r.Horas
}

// Nota_t é a Nota Gobrax oficial do motorista (premiação da API): o mês do dia analisado e o mês anterior
type Nota_t struct {
	Mes      string // AAAA-MM
	Atual    *float64
	Anterior *float64
	Parcial  bool // o mês ainda não terminou
}

// ChaveMotorista junta o mesmo nome escrito com espaços ou maiúsculas diferentes (diário x premiação)
func ChaveMotorista(nome string) string {
	return strings.ToUpper(strings.Join(strings.Fields(nome), " "))
}

// MontarResumos monta a análise de cada motorista que rodou no dia, comparada com o dia anterior em que ele rodou.
// regs traz o dia e os dias anteriores (DiasComparacao); um motorista que rodou mais de um veículo no dia entra uma
// vez só, com os valores juntos.
// notas vem pela ChaveMotorista; motorista sem nota recebe "ainda não calculada".
func MontarResumos(dia string, regs []diario.Registro_t, notas map[string]Nota_t) []Resumo_t {
	hoje := map[string][]diario.Registro_t{}
	antes := map[string][]diario.Registro_t{}
	for _, r := range regs {
		nome := strings.TrimSpace(r.Motorista)
		if nome == "" || !r.Rodou {
			continue
		}
		switch {
		case r.Dia == dia:
			hoje[nome] = append(hoje[nome], r)
		case r.Dia < dia:
			antes[nome] = append(antes[nome], r)
		}
	}
	// a comparação é com o dia anterior: o mais recente antes do dia em que o motorista rodou
	for nome, rs := range antes {
		ultimo := ""
		for _, r := range rs {
			if r.Dia > ultimo {
				ultimo = r.Dia
			}
		}
		so := rs[:0:0]
		for _, r := range rs {
			if r.Dia == ultimo {
				so = append(so, r)
			}
		}
		antes[nome] = so
	}
	nomes := make([]string, 0, len(hoje))
	for n := range hoje {
		nomes = append(nomes, n)
	}
	sort.Strings(nomes)

	out := make([]Resumo_t, 0, len(nomes))
	for _, nome := range nomes {
		out = append(out, montar(nome, dia, hoje[nome], antes[nome], notas[ChaveMotorista(nome)]))
	}
	return out
}

// comparacao_t é um indicador no dia contra o dia anterior (media = valor do dia anterior)
type comparacao_t struct {
	ind         indicador_t
	hoje, media *float64
}

// diferenca positiva = melhorou, negativa = piorou, 0 = estável ou sem comparação
func (c comparacao_t) sentido() int {
	if c.hoje == nil || c.media == nil {
		return 0
	}
	d := (*c.hoje - *c.media) * float64(c.ind.melhor)
	switch {
	case d >= c.ind.estavel:
		return 1
	case d <= -c.ind.estavel:
		return -1
	}
	return 0
}

func (c comparacao_t) texto() string {
	if c.hoje == nil {
		return "sem dado no dia"
	}
	v := num(*c.hoje, c.ind.casas) + c.ind.unidade
	if c.media == nil {
		return v
	}
	dif := num(math.Abs(*c.hoje-*c.media), c.ind.casas)
	if c.ind.unidade == "%" {
		dif += " pts"
	} else {
		dif += c.ind.unidade
	}
	switch c.sentido() {
	case 1:
		return v + " ✅ melhorou " + dif
	case -1:
		return v + " ⚠️ piorou " + dif
	}
	return v + " ➖ estável"
}

func montar(nome, dia string, hoje, antes []diario.Registro_t, nota Nota_t) Resumo_t {
	r := Resumo_t{Motorista: nome, Dia: dia, Km: soma(hoje, km), Kml: agregar(hoje, kml), Nota: nota.Atual}
	comps := []comparacao_t{{kml, r.Kml, agregar(antes, kml)}}
	for _, ind := range indicadores {
		comps = append(comps, comparacao_t{ind, agregar(hoje, ind), agregar(antes, ind)})
	}
	anterior := ""
	if len(antes) > 0 {
		anterior = antes[0].Dia
	}
	r.DiaAnterior = anterior
	r.Melhorou, r.Piorou, r.Comparados = contar(comps)
	r.Evolucao = evolucao(r.Melhorou, r.Piorou, r.Comparados, anterior)
	r.Dica = dicaDoDia(comps[1:], anterior)
	if p := piorIndicador(comps[1:]); p != nil && len(antes) > 0 {
		r.Foco, r.FocoRotulo, r.FocoValor, r.FocoMedia = p.ind.chave, p.ind.rotulo, p.hoje, p.media
	}
	r.Variaveis = map[string]string{
		"1":  primeiroNome(nome),
		"2":  dataBR(dia),
		"3":  comparadoCom(anterior),
		"4":  textoKm(r.Km),
		"13": textoNota(nota),
		"14": r.Evolucao,
		"15": r.Dica,
	}
	for _, c := range comps {
		r.Variaveis[c.ind.variavel] = c.texto()
	}
	return r
}

// contar: em quantos indicadores o motorista melhorou e em quantos piorou, entre os que deu para comparar
func contar(comps []comparacao_t) (melhorou, piorou, comparados int) {
	for _, c := range comps {
		if c.hoje == nil || c.media == nil {
			continue
		}
		comparados++
		switch c.sentido() {
		case 1:
			melhorou++
		case -1:
			piorou++
		}
	}
	return
}

func dataBR(dia string) string { return dia[8:10] + "/" + dia[5:7] }

// comparadoCom: o pedaço do título que diz com que dia a análise compara
func comparadoCom(anterior string) string {
	if anterior == "" {
		return ""
	}
	return " (comparado com " + dataBR(anterior) + ")"
}

// evolucao resume o dia em uma frase
func evolucao(melhorou, piorou, comparados int, anterior string) string {
	if anterior == "" {
		return "Ainda não temos um dia anterior seu para comparar. A partir do próximo dia você acompanha a sua evolução."
	}
	base := fmt.Sprintf("Em relação ao dia %s, você melhorou em %d e piorou em %d dos %d indicadores.", dataBR(anterior), melhorou, piorou, comparados)
	switch {
	case comparados == 0:
		return "Neste dia não houve indicadores registrados para comparar com o dia " + dataBR(anterior) + "."
	case melhorou > piorou:
		return "Você evoluiu! " + base
	case piorou > melhorou:
		return "Você regrediu um pouco. " + base
	}
	return "Você manteve o nível. " + base
}

// piorIndicador: o percentual que mais piorou em relação ao dia anterior (nil = nenhum piorou o bastante)
func piorIndicador(comps []comparacao_t) *comparacao_t {
	var pior *comparacao_t
	maior := 0.0
	for i := range comps {
		c := &comps[i]
		if c.hoje == nil || c.media == nil {
			continue
		}
		p := (*c.media - *c.hoje) * float64(c.ind.melhor) // positivo = piorou
		if p >= pioraMinima && p > maior {
			pior, maior = c, p
		}
	}
	return pior
}

// dicaDoDia: o que melhorar, a partir do indicador que mais piorou
func dicaDoDia(comps []comparacao_t, anterior string) string {
	if anterior == "" {
		return "mantenha a rotação no início da faixa verde e desligue o motor nas esperas. No próximo dia você recebe a comparação com este."
	}
	pior := piorIndicador(comps)
	if pior == nil {
		return "nenhum indicador piorou em relação ao dia " + dataBR(anterior) + ". Continue assim!"
	}
	return fmt.Sprintf("%s ficou em %s%% (no dia %s: %s%%). Dica: %s", pior.ind.rotulo, num(*pior.hoje, 1), dataBR(anterior), num(*pior.media, 1), pior.ind.dica)
}

func textoNota(n Nota_t) string {
	if n.Atual == nil {
		return "ainda não calculada"
	}
	t := num(*n.Atual, 1) + " em " + nomeMes(n.Mes)
	if n.Parcial {
		t += " (parcial)"
	}
	if n.Anterior == nil {
		return t
	}
	ant := nomeMes(MesAnterior(n.Mes))
	d := *n.Atual - *n.Anterior
	switch {
	case d >= 0.5:
		return t + " ✅ subiu " + num(d, 1) + " pontos em relação a " + ant
	case d <= -0.5:
		return t + " ⚠️ caiu " + num(-d, 1) + " pontos em relação a " + ant
	}
	return t + " ➖ igual a " + ant
}

var meses = []string{"janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}

func nomeMes(mes string) string {
	t, err := time.Parse("2006-01", mes)
	if err != nil {
		return mes
	}
	return meses[t.Month()-1]
}

// MesAnterior: "2026-01" vira "2025-12"
func MesAnterior(mes string) string {
	t, err := time.Parse("2006-01", mes)
	if err != nil {
		return ""
	}
	return t.AddDate(0, -1, 0).Format("2006-01")
}

// agregar junta os registros como o painel: razão (km/l) ou média ponderada; sem peso em nenhum registro, média simples
func agregar(regs []diario.Registro_t, ind indicador_t) *float64 {
	if ind.razao {
		var k, l float64
		for _, r := range regs {
			if r.Km != nil && r.Litros != nil && *r.Litros > 0 {
				k += *r.Km
				l += *r.Litros
			}
		}
		if l == 0 {
			return nil
		}
		x := utils.Arred2(k / l)
		return &x
	}
	var soma, pesos, simples float64
	n := 0
	for _, r := range regs {
		v := ind.valor(r)
		if v == nil {
			continue
		}
		n++
		simples += *v
		if p := ind.peso(r); p != nil && *p > 0 {
			soma += *v * *p
			pesos += *p
		}
	}
	if n == 0 {
		return nil
	}
	x := simples / float64(n)
	if pesos > 0 {
		x = soma / pesos
	}
	x = utils.Arred2(x)
	return &x
}

func soma(regs []diario.Registro_t, campo func(diario.Registro_t) *float64) *float64 {
	var s float64
	tem := false
	for _, r := range regs {
		if v := campo(r); v != nil {
			s += *v
			tem = true
		}
	}
	if !tem {
		return nil
	}
	s = utils.Arred2(s)
	return &s
}

// primeiroNome: a API costuma mandar o nome todo em maiúsculas ("JOAO DA SILVA" vira "Joao")
func primeiroNome(nome string) string {
	p := strings.Fields(nome)
	if len(p) == 0 {
		return "motorista"
	}
	r := []rune(strings.ToLower(p[0]))
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func textoKm(v *float64) string {
	if v == nil {
		return "não informado"
	}
	return num(*v, 0) + " km"
}

// num escreve no padrão brasileiro: 1.234,5
func num(v float64, casas int) string {
	s := fmt.Sprintf("%.*f", casas, math.Abs(v))
	inteira, dec, _ := strings.Cut(s, ".")
	var b strings.Builder
	for i, c := range inteira {
		if i > 0 && (len(inteira)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	if dec != "" {
		b.WriteString("," + dec)
	}
	if v < 0 && strings.Trim(s, "0.") != "" {
		return "-" + b.String()
	}
	return b.String()
}

// Texto preenche o modelo com as variáveis do motorista: é a mensagem pronta para copiar
func Texto(vars map[string]string) string {
	t := TextoTemplate
	for i := 1; i <= totalVariaveis; i++ {
		k := fmt.Sprint(i)
		t = strings.ReplaceAll(t, "{{"+k+"}}", vars[k])
	}
	return t
}
