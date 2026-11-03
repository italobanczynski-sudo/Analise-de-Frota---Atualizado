package hubspot

import (
	"regexp"
	"sort"
	"strings"

	"analisefrota/adapters/hubspotapi"
)

// Dados pessoais não saem do HubSpot (Política de Segurança da Gobrax, cláusula 5.1):
// os campos não são pedidos e o texto livre (notas, reuniões, ligações, assuntos) sai mascarado.
var (
	reEmail       = regexp.MustCompile(`[\w.+-]+@[\w-]+(\.[\w-]+)+`)
	reCNPJ        = regexp.MustCompile(`\b\d{2}\.?\d{3}\.?\d{3}/?\d{4}-?\d{2}\b`)
	reCPF         = regexp.MustCompile(`\b\d{3}\.?\d{3}\.?\d{3}-?\d{2}\b`)
	reTelefone    = regexp.MustCompile(`(\+?55\s?)?\(?\b\d{2}\)?\s?9?\d{4}[-\s]?\d{4}\b`)
	reTag         = regexp.MustCompile(`<[^>]+>`)
	reQuebra      = regexp.MustCompile(`(?i)<br\s*/?>|</p>|</li>`)
	reEspacos     = regexp.MustCompile(`[ \t]+`)
	reLinhas      = regexp.MustCompile(`\n\s*\n+`)
	reSufixo      = regexp.MustCompile(`\b(ltda|eireli|epp|me|s a|sa)\b`)
	reSufixoBusca = regexp.MustCompile(`(?i)\b(ltda|eireli|epp|me|s/?a)\b\.?`)
	reSeparador   = regexp.MustCompile(`[.,/\-]`)
)

// acentos do português (nomes de empresa)
var semAcento = strings.NewReplacer("á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a", "é", "e", "ê", "e", "è", "e", "í", "i", "ì", "i", "ó", "o", "ô", "o", "õ", "o", "ò", "o", "ú", "u", "ü", "u", "ç", "c", "ñ", "n")

var camposTexto = []string{"hs_note_body", "hs_call_body", "hs_meeting_body", "hs_internal_meeting_notes", "hs_email_subject", "hs_call_title", "hs_meeting_title", "subject"}
var camposCorpo = []string{"hs_call_body", "hs_meeting_body", "hs_internal_meeting_notes", "hs_note_body"}
var camposPessoais = []string{"email", "phone", "mobilephone", "cnpj_texto"}

func semPII(t string) string {
	t = reEmail.ReplaceAllString(t, "[e-mail]")
	t = reCNPJ.ReplaceAllString(t, "[CNPJ]")
	t = reCPF.ReplaceAllString(t, "[CPF]")
	return reTelefone.ReplaceAllString(t, "[telefone]")
}

// semHTML deixa o corpo das reuniões, ligações e notas em texto curto (vira o "o que foi combinado" no painel)
func semHTML(h string) string {
	h = reQuebra.ReplaceAllString(h, "\n")
	h = reTag.ReplaceAllString(h, "")
	h = strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">").Replace(h)
	h = reLinhas.ReplaceAllString(reEspacos.ReplaceAllString(h, " "), "\n")
	h = strings.TrimSpace(h)
	if r := []rune(h); len(r) > 800 {
		h = string(r[:800])
	}
	return h
}

// limpar tira os campos pessoais e mascara o texto livre
func limpar(lista []hubspotapi.Objeto_t) []hubspotapi.Objeto_t {
	for _, o := range lista {
		for _, c := range camposPessoais {
			delete(o.Properties, c)
		}
		for _, c := range camposCorpo {
			if v, ok := o.Properties[c]; ok {
				o.Properties[c] = semHTML(v)
			}
		}
		for _, c := range camposTexto {
			if v, ok := o.Properties[c]; ok {
				o.Properties[c] = semPII(v)
			}
		}
	}
	return lista
}

// chaveNome compara nomes de empresa ignorando caixa, acento, pontuação e LTDA/S.A. (igual ao ckey do painel)
func chaveNome(v string) string {
	s := semAcento.Replace(strings.ToLower(v))
	s = strings.Join(strings.Fields(s), " ")
	s = reSeparador.ReplaceAllString(s, " ")
	s = reSufixo.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}

// textoBusca é o nome sem LTDA/S.A. para a busca de empresas
func textoBusca(nome string) string {
	s := strings.TrimSpace(reSufixoBusca.ReplaceAllString(nome, ""))
	if r := []rune(s); len(r) > 100 {
		s = string(r[:100])
	}
	return s
}

// maisRecentes ordena do mais novo para o mais antigo pelo campo de data (texto ISO)
func maisRecentes(lista []hubspotapi.Objeto_t, campo string) {
	sort.SliceStable(lista, func(i, j int) bool { return lista[i].Properties[campo] > lista[j].Properties[campo] })
}

func primeiros(lista []hubspotapi.Objeto_t, n int) []hubspotapi.Objeto_t {
	if len(lista) > n {
		return lista[:n]
	}
	return lista
}
