# Análise de Frota (Gobrax)

Painel estático (HTML + JS) com endereço no **GitHub Pages**: https://italobanczynski-sudo.github.io/Analise-de-Frota---Atualizado/ . A parte online (login, base compartilhada, busca automática e rastro da frota) fica no servidor da Gobrax, em https://analisefrota.gobrax.com.br: quando ele está no ar, o endereço do GitHub leva direto para lá (`?local=1` fica no GitHub). Implantação em [servidor/README.md](servidor/README.md#implantação). O mesmo `index.html` roda em:

| Onde | Dados (clientes e planilhas) |
|---|---|
| **Painel no Claude** (Artifact): https://claude.ai/artifact/JtQke6GEAqzz8ohdP653dS | Base compartilhada: todos que têm acesso veem os mesmos dados, ao vivo. |
| **GitHub Pages** (endereço oficial) | Só no navegador de cada pessoa (localStorage). O navegador chama a API Gobrax direto com o token salvo; a Frota ao vivo atualiza de 5 em 5 min e o rastro do dia existe enquanto a página está aberta. |

## Estrutura do projeto

| Pasta ou arquivo | O que é |
|---|---|
| `index.html` | O painel (telas, gráficos e regras), publicado no GitHub Pages e no Claude |
| `bibliotecas/` | Biblioteca de terceiros usada pelo painel (leitor de Excel) |
| `documentacao/` | Documentação da API Gobrax usada pelo painel |
| `api/` | Scripts em Node que alimentam o painel no Claude até a fase 2: `puxar.mjs` (API Gobrax) e `hubspot.mjs`. Tokens em `busca-de-dados/.env` e lista em `busca-de-dados/clientes.json`, ambos fora do Git |
| `servidor/` | Plataforma em Go: login, perfis, busca automática, banco e API. Veja [servidor/README.md](servidor/README.md) |
| `.vscode/`, `.editorconfig`, `.gitattributes` | Padrão do projeto: ícones coloridos, formatação igual em qualquer editor e fim de linha LF |

## Painel no Claude
- As planilhas ficam na coleção `planilhas` do banco do Artifact: um documento por tipo (motoristas/veículos), cliente e dia, dividido em partes (`~p1`, `~p2`…) quando passa de 200 KB.
- Para publicar, o leitor de Excel é carregado do cdnjs em vez de `bibliotecas/` (troque o `src` do script ao gerar a cópia publicada).
- Acesso pelo menu Compartilhar do Artifact: Colaborador (Contributor) anexa e apaga planilhas; Leitor (Viewer) só vê.

## API Gobrax
Documentação completa em [documentacao/API.md](documentacao/API.md) e na aba **API Gobrax** do painel (`#/api`).

Para puxar os dados fora do navegador (o painel no Claude bloqueia a chamada): token em `busca-de-dados/.env`, clientes em `busca-de-dados/clientes.json` e `node busca-de-dados/puxar.mjs`.

No perfil do cliente, card "Buscar dados pela API Gobrax" (base `https://gateway-v3.gobrax.com.br:8889/api/`, `Authorization: Bearer <token>`):
- `v1/vehicle-statistics`: uma chamada por dia (dia de Brasília convertido para UTC), entra como planilha de veículos.
- `v1/vehicle-performance`: mesma janela diária, juntada às estatísticas pela placa: faixa verde (greenRange), embalo (leverage), motor ocioso (idle), pressão do pedal crítica (pedalPressureOnHigh), acima do verde (powerRange), excesso de velocidade em tempo (speeding), piloto automático (cruiseControl), freio motor (engineBrake), horas em movimento e o nome do motorista vinculado (CPF descartado). Se falhar, as estatísticas entram assim mesmo.
- `v2/vehicle-odometer`: último odômetro dos últimos 30 dias, card "Odômetro e comunicação" no fim de Desenvolvimento da frota.
- `v2/driversOverview`: nota, km e bonificação por mês, tela Premiação. O CPF devolvido é descartado antes de gravar.
- O token de cada cliente fica só no `localStorage` de quem o digitou (`gbx-api-token:<cliente>`), nunca na base compartilhada nem no código.
- Limites da API: 10 chamadas/s e no máximo 31 dias por consulta.

## Plataforma com login (servidor em Go)
A pasta [servidor/](servidor/README.md) tem a plataforma em Go (padrão Gobrax). Ela busca sozinha os dados da API Gobrax a cada hora, guarda tudo no Postgres (tokens cifrados) e entrega pela API HTTP e por telas com login, perfis (administrador, colaborador e leitor) e gestão de usuários. Substitui o `busca-de-dados/puxar.mjs`. O painel `index.html` ainda lê os dados pelo Artifact até a fase 2.

Ela não roda no GitHub (o Pages só hospeda páginas estáticas): fica no servidor da Gobrax. Começo rápido (precisa só do Docker): `cd servidor`, copie `.env.example` para `.env`, rode `docker compose up -d --build` e abra http://localhost:8080. Instalação, variáveis, perfis, API, testes, implantação e backup estão no [README do servidor](servidor/README.md).

## Publicar no GitHub Pages
1. Envie estes arquivos na raiz do repositório.
2. Settings → Pages → Source: "Deploy from a branch" → Branch `main` / pasta `/ (root)`.
3. Cada envio para a `main` atualiza o site em 1 a 2 minutos (aba Actions → pages build and deployment).

Para trazer os dados da plataforma antiga: entre nela, abra **Configurações → Levar os dados para outro lugar → Exportar tudo** e, no GitHub Pages, use **Importar arquivo** no mesmo cartão. Os tokens não vão no arquivo: salve o de cada cliente de novo no perfil dele.

O Render deixou de ser usado em 05/11/2026 (`render.yaml` e a tarefa que mantinha o servidor acordado foram removidos).

## Testes
A cada envio para o GitHub rodam (aba Actions → Testes):
- **Painel:** `node testes/painel/rodar.js` abre todas as telas e abas num Chrome sem janela, com dados sintéticos (com e sem Nota Gobrax), e falha se alguma tela der erro ou mostrar NaN/undefined.
- **Servidor:** `go build`, `go vet` e `go test` em `servidor/`.

## Observações
- O repositório/Pages contém só o código; nenhuma planilha ou dado de cliente vai junto.
