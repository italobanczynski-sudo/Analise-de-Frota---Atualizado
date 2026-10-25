# Plataforma Análise de Frota: servidor (Go)

Busca os dados da API Gobrax de cada cliente, guarda no Postgres e entrega pela API HTTP e pelas telas da plataforma (login, clientes e usuários). Substitui o `api/puxar.mjs`. Segue o padrão Go da Gobrax (Clean Architecture: `adapters` → `modules` → `api` → `background`).

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
| Implantação | Docker Compose (+ nginx e Let's Encrypt) ou Railway |

## Estrutura

```
main.go        sobe tudo; também tem os comandos -criar-admin e -importar-clientes
environment/   única leitura de variáveis de ambiente (Config_t carregado uma vez)
adapters/      Postgres (Singleton), migrações SQL, cliente da API Gobrax, e-mail, log
common/        datas de Brasília, placas, limpeza de token, validação, criptografia
modules/       usuario, sessao, cliente, diario, odometro, premiacao, coleta (regras + repositórios)
api/           route.go (rotas e perfis), auth.go (filtros HTTP), controller (entrada), middleware (casos de uso)
api/web/       telas: paginas/*.html e estatico/ (CSS, JS, imagens)
background/    tarefas automáticas
nginx/         analisefrota.conf (HTTPS) e http-somente.conf (só para gerar o certificado)
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
4. Opcional: traga os clientes do script antigo (`api/clientes.json` e os tokens do `api/.env`):
   ```bash
   docker compose run --rm app -importar-clientes /importar/clientes.json -tokens /importar/.env
   ```

Parar: `docker compose down` (os dados ficam no volume `dados`; `docker compose down -v` **apaga o banco**).

## Variáveis de ambiente

| Variável | Obrigatória | Para quê |
|---|---|---|
| `POSTGRES_PASSWORD` | sim (Compose) | Senha do Postgres do Compose |
| `DATABASE_URL` | sim | Montada pelo Compose; no Railway, vem do Postgres do projeto |
| `CHAVE_TOKENS` | sim | 32 bytes em base64 (`openssl rand -base64 32`). Cifra os tokens dos clientes e assina os códigos por e-mail. Trocar a chave invalida os tokens salvos e os códigos pendentes |
| `URL_PUBLICA` | não | Começo dos links de convite. Se começar com `https://`, o cookie de login é `Secure` (veja "Login só por HTTPS"). Padrão `http://localhost:8080` |
| `ORIGENS_PERMITIDAS` | não | Origens liberadas no CORS, separadas por vírgula |
| `PROXIES_CONFIAVEIS` | não | IP/CIDR dos proxies na frente do servidor. O Compose já confia na rede interna do Docker. Vazio = IP direto da conexão |
| `SMTP_HOST`, `SMTP_PORTA`, `SMTP_USUARIO`, `SMTP_SENHA`, `SMTP_REMETENTE` | não | Envio dos códigos de acesso. Sem `SMTP_HOST`, nenhum e-mail sai e o código aparece só no log (use só no computador) |
| `SESSAO_HORAS` (12), `CONVITE_HORAS` (72) | não | Duração do login e do link de convite |
| `COLETA_AUTOMATICA` (true), `COLETA_INTERVALO_MIN` (60), `COLETA_DIAS` (30) | não | Busca automática |
| `GOBRAX_API_BASE`, `GOBRAX_INTERVALO_MS` (120) | não | API Gobrax e pausa entre chamadas |
| `PORTA` (8080), `PORTA_LOCAL` (8080, só Compose), `NGINX_SITE` (só Compose) | não | Portas e site do nginx |

## Funcionalidades

- **Login** por e-mail e senha; **primeiro acesso** e **esqueci minha senha** por código de 6 dígitos enviado por e-mail; **link de convite** gerado pelo administrador.
- **Clientes**: lista com a situação da última busca; administrador cadastra e troca o token; colaborador e administrador pedem uma busca na hora.
- **Usuários** (só administrador): convidar, trocar perfil, desativar/reativar, gerar novo link de convite ou de senha.
- **Busca automática** (`v1/vehicle-statistics`, `v1/vehicle-performance`, `v2/vehicle-odometer`, `v2/driversOverview`): logo que sobe e a cada `COLETA_INTERVALO_MIN`. Na primeira vez, os últimos `COLETA_DIAS`; depois, do último dia buscado (refeito) até hoje. 404 é dia sem dado; falha pontual pula o dia sem apagar o que já havia; três falhas seguidas desligam aquele tipo na rodada.
- **Dados pela API** (registros, odômetros, premiações). As telas da plataforma ainda não mostram esses dados: o painel `index.html` continua lendo do Artifact (fase 2).

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
| `POST /api/v1/clientes/:id/coleta` | colaborador, admin | Busca agora, em segundo plano (202; 409 se já estiver buscando) |
| `POST /api/v1/clientes` `{nome, hubspot_id, token}` | admin | Cadastra cliente (token opcional) |
| `PUT /api/v1/clientes/:id/token` `{token}` | admin | Troca o token |
| `GET /api/v1/usuarios` | admin | Pessoas cadastradas |
| `POST /api/v1/usuarios` `{nome, email, papel}` | admin | Cadastra e devolve o link de convite |
| `POST /api/v1/usuarios/:id/convite` | admin | Novo link (o anterior deixa de valer) |
| `PUT /api/v1/usuarios/:id` `{papel, ativo}` | admin | Troca perfil / ativa / desativa |

A documentação da API Gobrax (externa) está em [../docs/API.md](../docs/API.md).

## Banco e migrações

As migrações ficam em `adapters/migrations/*.sql` e rodam sozinhas ao subir, em ordem, uma vez cada (controle em `schema_migrations`). Tabelas: `clientes`, `registros_diarios`, `odometros`, `premiacoes`, `coletas`, `usuarios`, `convites`, `sessoes`, `codigos_acesso`. O CPF que a API Gobrax devolve nunca é lido (os structs nem declaram o campo); só o nome do motorista é guardado.

Para mudar o banco, crie um arquivo novo (`004_...sql`); nunca edite um que já rodou.

## Testes

```bash
docker compose --profile testes run --rm testes   # unitários (go test ./...)
bash testes/rodar-e2e.sh                          # ponta a ponta: banco descartável, não manda e-mail
```

O ponta a ponta sobe um projeto Docker separado (`analisefrota-e2e`, porta 18080), cria o admin, roda `testes/e2e.sh` e apaga tudo ao final. Ele cobre telas, cabeçalhos de segurança, CORS, CSRF, convite, login, senha errada, trava por e-mail e por IP, logout, sessão vencida, troca de perfil, desativação, último admin, código por e-mail e a matriz de permissões dos três perfis.

**Resultado em 25/10/2026:** 12 pacotes Go com testes passando (`go vet` e `staticcheck` limpos, `deadcode` sem código morto) e 93 de 93 verificações ponta a ponta passando. O `govulncheck` não encontrou nenhuma vulnerabilidade alcançável.

## Implantação

### Servidor próprio (Docker + nginx + HTTPS)
Pré-requisitos: DNS de `analisefrota.gobrax.com.br` apontando para a máquina, portas 80 e 443 abertas, `.env` preenchido com `URL_PUBLICA=https://analisefrota.gobrax.com.br`.

```powershell
./setup-producao.ps1
```

O script sobe tudo com o nginx em HTTP (`NGINX_SITE=http-somente`), gera o certificado com o Certbot, troca o nginx para HTTPS (`analisefrota.conf`, o padrão) e confere `/saude`. Depois disso, para atualizar: `docker compose --profile producao up -d --build`.

Renovar o certificado (a cada ~60 dias; agende): `docker compose --profile certbot run --rm certbot renew` e depois `docker compose --profile producao restart nginx`.

### Railway
`railway.json` usa o `Dockerfile` e confere `/saude`. Configure no Railway: `DATABASE_URL` (do Postgres do projeto), `CHAVE_TOKENS`, `URL_PUBLICA`, `ORIGENS_PERMITIDAS`, SMTP e `PROXIES_CONFIAVEIS` com a faixa do proxy do Railway. **Não verificado nesta auditoria** (sem acesso à conta). Sem `PROXIES_CONFIAVEIS`, a trava por IP vale para todos juntos.

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
| nginx reiniciando sem parar | `analisefrota.conf` sem certificado. Rode `setup-producao.ps1` ou suba com `NGINX_SITE=http-somente` |
| "muitas tentativas" para todo mundo | Proxy na frente sem `PROXIES_CONFIAVEIS`, ou seja, todos com o mesmo IP. Configure a variável; a trava some sozinha em 15 min ou ao reiniciar o app |
| Código por e-mail não chega | Veja `docker compose logs app`: "SMTP não configurado" ou erro de login no SMTP (senha de app do Gmail) |
| "erro interno; veja o log do servidor" | Detalhe em `docker compose logs app` |
| Busca com erro "token" | Token do cliente inválido ou vencido: troque em `PUT /clientes/:id/token` |

## Limitações conhecidas

- A trava de tentativas fica na memória: zera ao reiniciar e não é compartilhada entre réplicas (rode uma réplica só).
- A sessão tem duração fixa (`SESSAO_HORAS`), sem renovação automática enquanto a pessoa usa.
- Não há registro de auditoria em tabela: só no log (`usuarios: X alterou Y`).
- As telas da plataforma mostram clientes e usuários, mas ainda não os gráficos e indicadores (estes continuam no painel `index.html`).
- A fonte Inter vem do Google Fonts (liberado na CSP); sem internet, cai para a fonte do sistema.

## Próximas fases

2. Telas de indicadores lendo do servidor (substituir o painel no Artifact).
3. HubSpot no servidor.
4. Telas em React (padrão Gobrax).
