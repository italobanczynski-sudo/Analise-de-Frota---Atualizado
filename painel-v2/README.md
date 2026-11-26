# Painel novo · Análise de frota Gobrax

Painel com a identidade visual Gobrax, num único HTML (JS puro, sem framework). Só lê e mostra: não tem formulário de preenchimento.

## Onde roda e de onde vêm os dados

| Onde | Dados |
|---|---|
| Servidor, em `/painel` (com login) | Os reais: o que o servidor busca de hora em hora na API Gobrax |
| Arquivo aberto direto ou GitHub Pages | Simulados (selo **DADOS SIMULADOS** no topo) |

O painel descobre sozinho em qual dos dois está: se `/api/v1/auth/eu` responde, está na plataforma.

No servidor, ele lê:

| Rota | Para quê |
|---|---|
| `/api/v1/painel/dados` | Dia a dia de cada veículo (último ano), motoristas vinculados, odômetro e premiação |
| `/api/v1/clientes/:id/ao-vivo` | Posição agora e rastro de hoje (mapa e trajeto do dia), atualizado a cada 2 minutos |
| `/api/v1/clientes/:id/mensagens/enviadas` | Evolução da nota depois das mensagens marcadas como enviadas |
| `/api/v1/clientes/:id/condutores` | Telefone do condutor para o botão WhatsApp (só CS e administrador) |
| `/mapa/:z/:x/:y` | Mapa-base do OpenStreetMap, servido pelo próprio servidor |

O painel anterior continua em `/painel/classico` (importação de planilhas, fichas e combinados).

## Montar

```
python painel-v2/build.py      # gera painel-v2/dist/analise-de-frota.html
node testes/painel-v2/rodar.js # varre todas as telas nos dois modos num Chrome sem janela
```

O `dist/` vai para o Git: o Dockerfile copia esse arquivo para dentro do servidor (`api/web/painel/novo.html`). O CI falha se alguém mexer em `src/` e esquecer de remontar.

`node testes/painel-v2/rodar.js --fotos <pasta>` tira capturas das telas principais no modo plataforma.

## Estrutura

- `src/data.js`    adaptador `Gobrax.api` (servidor ou simulação), agregações e regras de negócio
- `src/ui.js`      formatadores, gráfico de linha, medidores, overlays, estado
- `src/charts.js`  radar, dispersão, barras, calendário, catálogo de indicadores
- `src/maps.js`    mapa Leaflet, agrupamento de veículos, trajeto do dia
- `src/views_a.js` Visão geral, Saúde do cliente, Estudo
- `src/views_b.js` Frota e Motoristas
- `src/views_d.js` gráficos dos Motoristas, Mensagens, Plano de ação
- `src/views_c.js` Regras e integração
- `src/app.js`     rotas, ações, conta, sino de notificações
- `src/style.css`  Design System Gobrax
- `vendor/`        Leaflet 1.9.4 embutido

## Regras de cálculo que valem para todas as telas

- Cada linha de Motoristas é um motorista (vínculos da API, `motoristas` de `/painel/dados`), com o dado só do tempo em que ele estava vinculado a cada veículo; quem só revezou um veículo também aparece. Cliente sem motorista vinculado na API (e o modo simulado) continua com uma linha por veículo. A linha de base da economia é a mesma da frota.
- Na Frota, cada veículo mostra o motorista que mais o dirigiu no período.
- Nota Gobrax: a performance diária da API não traz nota (o servidor grava vazio). O painel usa a nota do mês de cada motorista na premiação Gobrax (v2/driversOverview) em cada dia em que ele dirigiu o veículo. Por isso a nota muda de mês em mês, não de dia em dia.
- Dia sem litros ou sem km fica fora do consumo. Indicador que a API não trouxe (nota, faixa verde, motor ocioso) fica fora da média daquele indicador; não vira zero.
- Registro com consumo fora de 1 a 7 km/l ou mais de 1.200 km no dia fica fora das médias (aba Frota › Qualidade do dado).
- Nota e premiação: quando a API traz a premiação oficial do mês (v2/driversOverview), a tela mostra ela; nos outros meses, uma estimativa pela regra da tela Regras e integração.
- Estudo do projeto: o início é o primeiro dia com dado dentro do último ano guardado no servidor. Unidades = veículos com dado; valor por unidade = premissa da tela Regras.
