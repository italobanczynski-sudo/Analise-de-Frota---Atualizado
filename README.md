# Análise de Frota (Gobrax)

A plataforma (login, painel, busca automática na API Gobrax e rastro da frota) está no **Render**: https://analise-de-frota.onrender.com . O painel fica em `/painel` (código em [painel-v2/](painel-v2/README.md)).

O painel clássico (`index.html` antigo, GitHub Pages e `/painel/classico`) saiu em 26/11/2026. O endereço do GitHub Pages e `/painel/classico` levam para `/painel`.

## Estrutura do projeto

| Pasta ou arquivo | O que é |
|---|---|
| `painel-v2/` | O painel, com a identidade visual Gobrax, servido em `/painel` na plataforma com os dados reais. Veja [painel-v2/README.md](painel-v2/README.md) |
| `index.html` | Só leva o endereço antigo do GitHub Pages para a plataforma |
| `robo-whatsapp/` | Robô em Node que roda num computador e, todo dia às 08:00, manda as Mensagens do dia pelo WhatsApp Web a cada motorista com telefone cadastrado e marca como enviada. Contra os Termos do WhatsApp: usar um número só para isso. Senha, sessão e logs fora do Git |
| `servidor/` | Plataforma em Go: login, perfis, busca automática, banco e API. Veja [servidor/README.md](servidor/README.md) |
| `render.yaml` | Implantação na nuvem no Render (plano grátis: https://analise-de-frota.onrender.com) |
| `.vscode/`, `.editorconfig`, `.gitattributes` | Padrão do projeto: ícones coloridos, formatação igual em qualquer editor e fim de linha LF |

## API Gobrax
A busca na API Gobrax fica no [servidor](servidor/README.md), que puxa os dados de cada cliente a cada hora (`v1/vehicle-statistics`, `v1/vehicle-performance`, `v2/vehicle-odometer` e `v2/driversOverview`; limites: 10 chamadas/s e no máximo 31 dias por consulta). O token de cada cliente fica cifrado no banco do servidor.

## Plataforma com login (servidor em Go)
A pasta [servidor/](servidor/README.md) tem a plataforma em Go (padrão Gobrax). Ela busca sozinha os dados da API Gobrax a cada hora, guarda tudo no Postgres (tokens cifrados) e entrega pela API HTTP e por telas com login, perfis (administrador, colaborador e leitor) e gestão de usuários.

Ela está no ar no Render. Começo rápido (precisa só do Docker): `cd servidor`, copie `.env.example` para `.env`, rode `docker compose up -d --build` e abra http://localhost:8080. Instalação, variáveis, perfis, API, testes, implantação e backup estão no [README do servidor](servidor/README.md).

## Hospedagem: atenção
- O banco grátis do Render expira 30 dias depois de criado (criado em 25/10/2026). Antes disso, passe o banco para um plano pago no painel do Render, senão clientes, tokens e histórico se perdem.
- O GitHub Actions chama `/saude` a cada 10 min para o servidor grátis não dormir (a busca automática de hora em hora só roda acordado).
- Para mudar depois para um servidor da Gobrax, o passo a passo (com a cópia dos dados do Render) está em [servidor/README.md](servidor/README.md#implantação).

## Testes
A cada envio para o GitHub rodam (aba Actions → Testes):
- **Painel:** confere se `painel-v2/dist/` foi remontado (`python painel-v2/build.py`) e roda `node testes/painel-v2/rodar.js`, que varre as telas em modo demonstração e com um servidor falso que responde no formato do servidor Go.
- **Servidor:** `go build`, `go vet` e `go test` em `servidor/`.

## Observações
- O repositório contém só o código; nenhuma planilha ou dado de cliente vai junto.
