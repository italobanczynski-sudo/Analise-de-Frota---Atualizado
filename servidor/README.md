# Plataforma Análise de Frota: servidor (Go)

Busca os dados da API Gobrax de cada cliente, guarda no Postgres e entrega pela API HTTP e pelas telas da plataforma (login, clientes e usuários). Substitui o `busca-de-dados/puxar.mjs`. Segue o padrão Go da Gobrax (Clean Architecture: `adapters` → `modules` → `api` → `background`).

## Tecnologias

| Parte | O que usa |
|---|---|
| Linguagem | Go 1.27 (mínimo 1.26, no `go.mod`) |
| HTTP | Gin 1.12, cookie de sessão HttpOnly |
| Banco | PostgreSQL 16 (sqlx + lib/pq), migrações SQL embutidas no binário |
| Tarefas | gocron (busca de hora em hora, limpeza de sessões) |
| Senhas e segredos | bcrypt (custo 12), AES-256-GCM para tokens de clientes, HMAC-SHA256 para códigos por e-mail |
| E-mail | SMTP (Gmail com senha de app, Workspace, Office 365...) |
| Telas | HTML, CSS e JS puros, embutidos no binário (`api/web`) |
| Implantação | Render (atual) ou Docker Compose (+ nginx e Let's Encrypt) |

## Estrutura

```
main.go        sobe tudo; também tem os comandos -criar-admin e -importar-clientes
environment/   única leitura de variáveis de ambiente (Config_t carregado uma vez)
adapters/      Postgres (Singleton), migrações SQL, cliente da API Gobrax, e-mail, log
common/        datas de Brasília, placas, limpeza de token, validação, criptografia
modules/       usuario, sessao, cliente, diario, odometro, premiacao, coleta, trajeto (regras + repositórios)
api/           route.go (rotas e perfis), auth.go (filtros HTTP), controller (entrada), middleware (casos de uso)
api/web/       telas: paginas/*.html, estatico/ (CSS, JS, imagens) e painel/ (cópia do index.html feita pelo Dockerfile)
background/    tarefas automáticas
implantacao/   colocar no ar: nginx/ (analisefrota.conf com HTTPS e http-somente.conf só para gerar o certificado),
               setup-producao.ps1 (HTTPS + e-mail) e copiar-dados-nuvem.sh (copia o banco local para a nuvem)
testes/        teste ponta a ponta (e2e.sh) e o ambiente descartável dele
```

## Rodar no computador

Precisa só do Docker Desktop.

1. Copie `.env.example` para `.env` e preencha (o `.env` fica fora do Git; nunca o cole em chat, e-mail ou commit).
2. Suba banco e servidor:
   ```bash
   docker compose up -d --build
   ```
   A plataforma abre em **http://localhost:8080** (só neste computador).
3. Crie o primeiro administrador. O comando mostra um link de convite (vale uma vez, por 72 h):
   ```bash
   docker compose run --rm app -criar-admin "Nome Sobrenome" -email pessoa@gobrax.com
   ```
   Troque o começo do link por `http://localhost:8080` se a `URL_PUBLICA` for a de produção.
   Se a pessoa já existe (convite perdido ou vencido, ninguém consegue entrar, e-mail fora do ar), gere um link novo:
   ```bash
   docker compose run --rm app -novo-convite pessoa@gobrax.com
   ```
4. Opcional: traga os clientes do script antigo (`busca-de-dados/clientes.json` e os tokens do `busca-de-dados/.env`):
   ```bash
   docker compose run --rm app -importar-clientes /importar/clientes.json -tokens /importar/.env
   ```

Parar: `docker compose down` (os dados ficam no volume `dados`; `docker compose down -v` **apaga o banco**).

## Variáveis de ambiente

| Variável | Obrigatória | Para quê |
|---|---|---|
| `POSTGRES_PASSWORD` | sim (Compose) | Senha do Postgres do Compose |
| `DATABASE_URL` | sim | Montada pelo Compose; no Render, vem do banco do projeto |
| `CHAVE_TOKENS` | sim | 32 bytes em base64 (`openssl rand -base64 32`). Cifra os tokens dos clientes e assina os códigos por e-mail. Trocar a chave invalida os tokens salvos e os códigos pendentes |
| `URL_PUBLICA` | não | Começo dos links de convite. Se começar com `https://`, o cookie de login é `Secure` (veja "Login só por HTTPS"). Padrão `http://localhost:8080` |
| `ORIGENS_PERMITIDAS` | não | Origens liberadas no CORS, separadas por vírgula |
| `PROXIES_CONFIAVEIS` | não | IP/CIDR dos proxies na frente do servidor. O Compose já confia na rede interna do Docker. Vazio = IP direto da conexão |
| `SMTP_HOST`, `SMTP_PORTA`, `SMTP_USUARIO`, `SMTP_SENHA`, `SMTP_REMETENTE` | não | Envio dos códigos de acesso. Sem `SMTP_HOST`, nenhum e-mail sai e o código aparece só no log (use só no computador) |
| `SESSAO_HORAS` (12), `CONVITE_HORAS` (72) | não | Duração do login e do link de convite |
| `COLETA_AUTOMATICA` (true), `COLETA_INTERVALO_MIN` (60), `COLETA_DIAS` (30), `COLETA_HOJE_MIN` (15) | não | Busca automática e atualização rápida do dia de hoje (0 desliga) |
| `RASTRO_INTERVALO_MIN` (5; 0 desliga), `RASTRO_DIAS` (7) | não | Rastro da frota ao vivo: de quanto em quanto tempo grava a posição de todos os veículos e por quantos dias guarda |
| `GOBRAX_API_BASE`, `GOBRAX_INTERVALO_MS` (120) | não | API Gobrax e pausa entre chamadas |
| `PORTA` (8080), `PORTA_LOCAL` (8080, só Compose), `NGINX_SITE` (só Compose) | não | Portas e site do nginx |

## Funcionalidades

- **Login** por e-mail e senha; **primeiro acesso** e **esqueci minha senha** por código de 6 dígitos enviado por e-mail; **link de convite** gerado pelo administrador.
- **Clientes**: lista com a situação da última busca; administrador cadastra e troca o token; colaborador e administrador pedem uma busca na hora.
- **Usuários** (só administrador): convidar, trocar perfil, desativar/reativar, gerar novo link de convite ou de senha.
- **Busca automática** (`v1/vehicle-statistics`, `v1/vehicle-performance`, `v2/vehicle-odometer`, `v2/driversOverview`): logo que sobe e a cada `COLETA_INTERVALO_MIN`. Na primeira vez, os últimos `COLETA_DIAS`; depois, do último dia buscado (refeito) até hoje. 404 é dia sem dado; falha pontual pula o dia sem apagar o que já havia; três falhas seguidas desligam aquele tipo na rodada. Entre uma busca e outra, a cada `COLETA_HOJE_MIN` o servidor refaz só o dia de hoje (estatísticas, performance e motoristas), para os números do painel andarem ao longo do dia; pula cliente em pausa pedida pela API, com erro ou nunca buscado. O painel relê os dados a cada 5 min.
- **Painel completo em `/painel`** (o mesmo `index.html` do GitHub Pages, com todas as telas): depois do login a pessoa cai nele. A ponte `api/web/estatico/painel-base.js` oferece ao painel a mesma "base compartilhada" do Artifact, guardada no Postgres (`base_documentos`: clientes, tratativas, configurações e planilhas importadas). Os dados que o servidor busca na API Gobrax chegam ao painel montados como as planilhas de veículos e motoristas, o odômetro e a premiação (último ano). Na planilha de motoristas, cada um tem só o tempo do vínculo dele com o veículo (`registros_motoristas`): quem ficou o dia todo usa a linha do veículo, e os vínculos parciais são buscados na API só nos trechos daquele motorista, como o painel faz. Isso gasta mais chamadas à performance nos dias com troca de motorista. O que alguém importou à mão para o mesmo cliente e dia vale mais. O leitor só lê; colaborador e admin gravam.
- **Mensagens do dia** (painel, menu Motoristas → Mensagens do dia): a análise de cada motorista que rodou no dia (km, média km/l e os indicadores da plataforma Gobrax: Início de Faixa Verde, Aproveitamento de Embalo, Motor Ligado Parado, Acelerando acima do Verde, Excesso de Velocidade, Piloto Automático e Pedal Crítico, cada um contra o dia anterior em que ele rodou; nota do mês contra o mês anterior, evolução e o que melhorar; com mais de um veículo no dia, cada percentual é ponderado pela própria base, como a plataforma), pronta para copiar e mandar pelo WhatsApp. Não há integração: quem manda marca como enviada. Cada marca guarda um retrato da análise (`mensagens_enviadas`), e o **Motorista individual** mostra os dias com mensagem e se ele melhorou, nos 7 dias seguintes, no ponto apontado.
- Em 25/10/2026 os 447 documentos da base do Artifact (2 clientes, 194 tratativas, diesel e 248 planilhas) foram copiados para o banco local.

## Login, sessões e perfis

| Perfil | Vê clientes e dados | Pede busca | Cadastra cliente / token | Cuida de usuários |
|---|---|---|---|---|
| `leitor` | sim | não | não | não |
| `colaborador` | sim | sim | não | não |
| `admin` | sim | sim | sim | sim |

- A senha tem de 10 a 72 caracteres, com letras e números; no banco fica só o hash bcrypt.
- O cookie `rg_sessao` é HttpOnly e SameSite=Strict, e vale `SESSAO_HORAS`. No banco fica só o SHA-256 do token. O perfil é relido do banco a cada pedido, então desativar a pessoa corta o acesso na hora.
- Trocar o perfil, desativar ou trocar a senha (por código ou por link) **encerra todos os logins abertos** da pessoa.
- Pedidos que mudam algo exigem o cabeçalho `X-Requisicao: plataforma`, que, somado ao SameSite=Strict, bloqueia CSRF.
- **Trava de tentativas** (memória do processo, 15 min): 5 erros do mesmo e-mail ou 20 do mesmo IP. Atrás de proxy, o IP real só é lido de `PROXIES_CONFIAVEIS`.
- E-mail inexistente recebe a mesma resposta que senha errada, e o pedido de código responde igual para e-mail cadastrado ou não.
- O código vale 15 min e uma vez só; aceita 5 tentativas e um novo pedido a cada 60 s.
- **Não há separação por empresa:** toda pessoa logada vê todos os clientes. É a regra atual (a plataforma é da equipe Gobrax).

### Login só por HTTPS
Com `URL_PUBLICA` em `https://`, o cookie é `Secure` e o navegador **descarta** o cookie numa página aberta por `http://` (exceto `http://localhost`). Antes desta auditoria, isso fazia a tela voltar para o login sem aviso. Agora a tela confere a sessão depois do login e explica o motivo. Em produção, abra sempre por `https://`.

## API

Respostas: `{"ok": true, "dados": ...}` ou `{"ok": false, "erro": "..."}`. Toda rota em `/api/v1` (menos `auth/*` de entrada) exige login; `POST`/`PUT` exigem `X-Requisicao: plataforma`.

| Método e rota | Perfil | O que faz |
|---|---|---|
| `GET /saude` | público | Servidor no ar |
| `GET /`, `/convite`, `/plataforma` | público | Telas |
| `POST /api/v1/auth/entrar` `{email, senha}` | público | Login (grava o cookie) |
| `POST /api/v1/auth/convite` `{token, senha}` | público | Aceita o convite e já entra |
| `POST /api/v1/auth/codigo` `{email}` | público | Envia o código por e-mail |
| `POST /api/v1/auth/codigo/confirmar` `{email, codigo, senha}` | público | Grava a senha nova e entra |
| `POST /api/v1/auth/sair` | público | Encerra a sessão |
| `GET /api/v1/auth/eu` | logado | Quem está logado |
| `GET /api/v1/clientes` | logado | Clientes e situação da busca (`tem_token`, nunca o token) |
| `GET /api/v1/clientes/:id/coleta` | logado | Situação: `nunca`, `rodando`, `ok` ou `erro` |
| `GET /api/v1/clientes/:id/registros?inicio=AAAA-MM-DD&fim=AAAA-MM-DD` | logado | Dia a dia de cada veículo |
| `GET /api/v1/clientes/:id/odometros` | logado | Último odômetro de cada veículo |
| `GET /api/v1/clientes/:id/premiacoes?mes=AAAA-MM` | logado | Premiação (sem `mes`: todos) |
| `GET /api/v1/clientes/:id/trajeto?placa=ABC1D23&dia=AAAA-MM-DD&de=0&ate=24` | logado | Trajeto do dia: um ponto a cada 5 min (horas de Brasília). `completo: false` = o tempo do pedido acabou; peça de novo, o que já veio fica no banco |
| `GET /api/v1/clientes/:id/ao-vivo?placas=ABC1D23,DEF4G56&desde=AAAA-MM-DD HH:MM:SS` | logado | Onde cada placa está agora (até 150): `movimento`, `ligado` (motor ligado parado), `desligado`, `sem_sinal` ou `nao_encontrado`, com a última posição, os minutos desde ela e a velocidade; e `rastro`, os pontos de ignição ligada de cada placa desde `desde` (UTC; sem ele, desde a meia-noite de Brasília). Cache de 90 s. `completo: false` = o tempo do pedido acabou; peça de novo |
| `GET /mapa/:z/:x/:y.png` | logado | Fundo do mapa do painel, buscado pelo servidor no OpenStreetMap com a identificação da plataforma (cache em memória) |
| `GET /painel` | logado (senão volta para `/`) | Painel completo |
| `GET /api/v1/painel/dados` | logado | Dados da API de cada cliente (último ano), para o painel |
| `GET /api/v1/base/:colecao` | logado | Documentos da base do painel |
| `PUT /api/v1/base/:colecao/:doc`, `POST /api/v1/base/:colecao`, `DELETE /api/v1/base/:colecao/:doc` | colaborador, admin | Grava, cria e apaga documentos da base do painel |
| `POST /api/v1/clientes/:id/coleta` | colaborador, admin | Busca agora, em segundo plano (202; 409 se já estiver buscando) |
| `POST /api/v1/clientes` `{nome, token}` | admin | Cadastra cliente (token opcional) |
| `PUT /api/v1/clientes/:id/token` `{token}` | admin | Troca o token |
| `GET /api/v1/usuarios` | admin | Pessoas cadastradas |
| `POST /api/v1/usuarios` `{nome, email, papel}` | admin | Cadastra e devolve o link de convite |
| `POST /api/v1/usuarios/:id/convite` | admin | Novo link (o anterior deixa de valer) |
| `PUT /api/v1/usuarios/:id` `{papel, ativo}` | admin | Troca perfil / ativa / desativa |
| `GET /api/v1/clientes/:id/mensagens?dia=AAAA-MM-DD` | logado | Mensagem pronta de cada motorista que rodou no dia (sem `dia`: ontem) e se já foi marcada como enviada |
| `GET /api/v1/clientes/:id/mensagens/acompanhamento?motorista=Nome` | logado | Dias com mensagem enviada ao motorista e se ele melhorou depois no ponto apontado |
| `GET /api/v1/clientes/:id/mensagens/enviadas` | logado | Por motorista: quantas mensagens foram marcadas como enviadas, a nota do mês em cada dia e os indicadores que ele foi orientado a melhorar (aba Motoristas › Mensagens enviadas) |
| `PUT /api/v1/clientes/:id/mensagens/enviada` `{dia, motorista}` | colaborador, admin | Marca a mensagem como enviada (guarda o retrato da análise) |
| `DELETE /api/v1/clientes/:id/mensagens/enviada?dia=&motorista=` | colaborador, admin | Desmarca |

A documentação da API Gobrax (externa) está em [../docs/API.md](../docs/API.md).

## Banco e migrações

As migrações ficam em `adapters/migrations/*.sql` e rodam sozinhas ao subir, em ordem, uma vez cada (controle em `schema_migrations`). Tabelas: `clientes`, `registros_diarios`, `registros_motoristas`, `odometros`, `premiacoes`, `coletas`, `posicoes_amostradas`, `usuarios`, `convites`, `sessoes`, `codigos_acesso`. O CPF que a API Gobrax devolve nunca é lido (os structs nem declaram o campo); só o nome do motorista é guardado.

Para mudar o banco, crie um arquivo novo (`004_...sql`); nunca edite um que já rodou.

## Testes

```bash
docker compose --profile testes run --rm testes   # unitários (go test ./...)
bash testes/rodar-e2e.sh                          # ponta a ponta: banco descartável, não manda e-mail
```

O ponta a ponta sobe um projeto Docker separado (`analisefrota-e2e`, porta 18080), cria o admin, roda `testes/e2e.sh` e apaga tudo ao final. Ele cobre telas, cabeçalhos de segurança, CORS, CSRF, convite, login, senha errada, trava por e-mail e por IP, logout, sessão vencida, troca de perfil, desativação, último admin, código por e-mail e a matriz de permissões dos três perfis.

**Resultado em 25/10/2026:** 12 pacotes Go com testes passando (`go vet` e `staticcheck` limpos, `deadcode` sem código morto) e 93 de 93 verificações ponta a ponta passando. O `govulncheck` não encontrou nenhuma vulnerabilidade alcançável.

## Implantação

### Hoje: Render
A plataforma está em https://analise-de-frota.onrender.com (`render.yaml` na raiz). Cada envio para a `main` é implantado sozinho.

### Para o futuro: servidor da Gobrax (Docker + nginx + HTTPS)

**O que a TI precisa providenciar (uma vez):**
1. Uma máquina Linux sempre ligada (2 vCPU, 2 GB de RAM e 20 GB de disco bastam), na rede da empresa ou na nuvem que a Gobrax já usa, com Docker, Docker Compose, Git e PowerShell (`pwsh`).
2. DNS: registro A de `analisefrota.gobrax.com.br` para o IP público da máquina.
3. Firewall: portas 80 e 443 abertas para a internet. A porta 8080 fica só na própria máquina.
4. Um e-mail com senha de app para os códigos de acesso (SMTP).

**Colocar no ar:**
1. `git clone https://github.com/italobanczynski-sudo/Analise-de-Frota---Atualizado.git` e `cd Analise-de-Frota---Atualizado/servidor`.
2. Copie `.env.example` para `.env` e preencha. Use a **mesma `CHAVE_TOKENS`** do Render (painel do Render → serviço → Environment), senão os tokens dos clientes trazidos de lá não abrem. Os segredos ficam só nesse arquivo, nunca em chat, e-mail ou commit.
3. Rode `pwsh ./implantacao/setup-producao.ps1` (sobe banco, app e nginx, gera o certificado e liga o HTTPS).
4. Traga os dados do Render: cole a "External Database URL" do banco `analise-de-frota-db` (Render → Connect) no arquivo `servidor/.banco-render` e rode `bash implantacao/trazer-dados-render.sh`. Ele guarda antes uma cópia do banco atual (`backup-antes-render.dump`). Depois, apague o `.banco-render`.
5. Atualização automática: agende `implantacao/atualizar.sh` a cada 5 min (`crontab -e`, a linha está no próprio script). Cada envio para a `main` no GitHub chega à plataforma em até 5 min.
6. Confira `https://analisefrota.gobrax.com.br`, entre e só então apague o serviço e o banco no painel do Render.

O certificado HTTPS se renova sozinho (`certbot-renovar` tenta a cada 12 h e o nginx relê a cada 6 h).

Pré-requisitos do script: DNS de `analisefrota.gobrax.com.br` apontando para a máquina, portas 80 e 443 abertas, `.env` preenchido com `URL_PUBLICA=https://analisefrota.gobrax.com.br`.

```powershell
./implantacao/setup-producao.ps1
```

O script sobe tudo com o nginx em HTTP (`NGINX_SITE=http-somente`), gera o certificado com o Certbot, troca o nginx para HTTPS (`analisefrota.conf`, o padrão) e confere `/saude`. Depois disso, para atualizar à mão: `docker compose --profile producao up -d --build` (ou deixe o `atualizar.sh` agendado).

## Backup e recuperação

```bash
docker compose exec -T db pg_dump -U frota -d frota -Fc > backup-AAAA-MM-DD.dump      # backup
docker compose exec -T db pg_restore -U frota -d frota --clean < backup-AAAA-MM-DD.dump # restaurar
```

Guarde a `CHAVE_TOKENS` junto do plano de backup, em um cofre de senhas e nunca no mesmo lugar do dump: sem ela, os tokens dos clientes no backup não podem ser decifrados. Não há backup automático configurado.

## Resolução de problemas

| Sintoma | Causa e saída |
|---|---|
| Login aceita a senha e volta para a tela de entrada | Página aberta por `http://` com `URL_PUBLICA` em `https://` (a tela agora explica isso). Abra por `https://` ou, no computador, use `http://localhost:8080` |
| nginx reiniciando sem parar | `analisefrota.conf` sem certificado. Rode `implantacao/setup-producao.ps1` ou suba com `NGINX_SITE=http-somente` |
| "muitas tentativas" para todo mundo | Proxy na frente sem `PROXIES_CONFIAVEIS`, ou seja, todos com o mesmo IP. Configure a variável; a trava some sozinha em 15 min ou ao reiniciar o app |
| Código por e-mail não chega | Veja `docker compose logs app`: "SMTP não configurado" ou erro de login no SMTP (senha de app do Gmail) |
| "erro interno; veja o log do servidor" | Detalhe em `docker compose logs app` |
| Busca com erro "token" | Token do cliente inválido ou vencido: troque em `PUT /clientes/:id/token` |

## Limitações conhecidas

- A trava de tentativas fica na memória: zera ao reiniciar e não é compartilhada entre réplicas (rode uma réplica só).
- A sessão tem duração fixa (`SESSAO_HORAS`), sem renovação automática enquanto a pessoa usa.
- Não há registro de auditoria em tabela: só no log (`usuarios: X alterou Y`).
- As telas da plataforma mostram clientes e usuários, mas ainda não os gráficos e indicadores (estes continuam no painel `index.html`).
- O painel em `/painel` usa scripts dentro do HTML, então a CSP dele libera `'unsafe-inline'`; as demais telas continuam com a CSP estrita.
- Apagar no painel um dado que veio da API (não importado à mão) não tem efeito: ele volta na próxima atualização.
- A imagem é construída a partir da raiz do repositório (`docker build -f servidor/Dockerfile .`).
- A fonte Inter vem do Google Fonts (liberado na CSP); sem internet, cai para a fonte do sistema.
- O trajeto do dia é amostrado: a API de posições só dá a última posição de cada intervalo, então o servidor pergunta janelas de 5 minutos (288 chamadas no dia). Essas chamadas entram na mesma fila da busca de hora em hora: com a busca rodando, o trajeto espera. A coleção de posições da Gobrax cita a porta 8888 (http); o servidor usa a mesma base dos outros endpoints (8889, https) até a Gobrax confirmar.
- O mapa do painel usa o Leaflet do cdnjs (liberado na CSP do `/painel`) e os quadros do OpenStreetMap.

## Próximas fases

2. Telas de indicadores lendo do servidor (substituir o painel no Artifact).
3. Telas em React (padrão Gobrax).
