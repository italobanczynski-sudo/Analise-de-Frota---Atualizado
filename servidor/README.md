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
api/web/       telas: paginas/*.html, estatico/ (CSS, JS, imagens) e painel/ (cópia do index.html feita pelo Dockerfile)
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
| `TWILIO_ACCOUNT_SID`, `TWILIO_API_KEY`, `TWILIO_API_SECRET`, `WHATSAPP_REMETENTE`, `WHATSAPP_TEMPLATE_SID` | não | WhatsApp dos motoristas (veja a seção abaixo). Faltando qualquer uma, nada é enviado e a mensagem aparece só no log |
| `WHATSAPP_AUTOMATICO` (false), `WHATSAPP_HORA` (7) | não | Envio diário automático e a hora de Brasília (0 a 23) |
| `PORTA` (8080), `PORTA_LOCAL` (8080, só Compose), `NGINX_SITE` (só Compose) | não | Portas e site do nginx |

## Funcionalidades

- **Login** por e-mail e senha; **primeiro acesso** e **esqueci minha senha** por código de 6 dígitos enviado por e-mail; **link de convite** gerado pelo administrador.
- **Clientes**: lista com a situação da última busca; administrador cadastra e troca o token; colaborador e administrador pedem uma busca na hora.
- **Usuários** (só administrador): convidar, trocar perfil, desativar/reativar, gerar novo link de convite ou de senha.
- **Busca automática** (`v1/vehicle-statistics`, `v1/vehicle-performance`, `v2/vehicle-odometer`, `v2/driversOverview`): logo que sobe e a cada `COLETA_INTERVALO_MIN`. Na primeira vez, os últimos `COLETA_DIAS`; depois, do último dia buscado (refeito) até hoje. 404 é dia sem dado; falha pontual pula o dia sem apagar o que já havia; três falhas seguidas desligam aquele tipo na rodada.
- **Painel completo em `/painel`** (o mesmo `index.html` do GitHub Pages, com todas as telas): depois do login a pessoa cai nele. A ponte `api/web/estatico/painel-base.js` oferece ao painel a mesma "base compartilhada" do Artifact, guardada no Postgres (`base_documentos`: clientes, tratativas, HubSpot, configurações e planilhas importadas). Os dados que o servidor busca na API Gobrax chegam ao painel montados como as planilhas de veículos e motoristas, o odômetro e a premiação (último ano). O que alguém importou à mão para o mesmo cliente e dia vale mais. O leitor só lê; colaborador e admin gravam.
- **WhatsApp dos motoristas** (aba WhatsApp, colaborador e administrador): toda manhã cada motorista com WhatsApp cadastrado recebe o resultado do dia anterior. Veja a seção [WhatsApp dos motoristas](#whatsapp-dos-motoristas).
- Em 25/10/2026 os 447 documentos da base do Artifact (2 clientes, 194 tratativas, HubSpot, diesel e 248 planilhas) foram copiados para o banco local.

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
| `GET /painel` | logado (senão volta para `/`) | Painel completo |
| `GET /api/v1/painel/dados` | logado | Dados da API de cada cliente (último ano), para o painel |
| `GET /api/v1/base/:colecao` | logado | Documentos da base do painel |
| `PUT /api/v1/base/:colecao/:doc`, `POST /api/v1/base/:colecao`, `DELETE /api/v1/base/:colecao/:doc` | colaborador, admin | Grava, cria e apaga documentos da base do painel |
| `POST /api/v1/clientes/:id/coleta` | colaborador, admin | Busca agora, em segundo plano (202; 409 se já estiver buscando) |
| `POST /api/v1/clientes` `{nome, hubspot_id, token}` | admin | Cadastra cliente (token opcional) |
| `PUT /api/v1/clientes/:id/token` `{token}` | admin | Troca o token |
| `GET /api/v1/usuarios` | admin | Pessoas cadastradas |
| `POST /api/v1/usuarios` `{nome, email, papel}` | admin | Cadastra e devolve o link de convite |
| `POST /api/v1/usuarios/:id/convite` | admin | Novo link (o anterior deixa de valer) |
| `PUT /api/v1/usuarios/:id` `{papel, ativo}` | admin | Troca perfil / ativa / desativa |
| `GET /api/v1/clientes/:id/whatsapp` | colaborador, admin | Configuração, telefones (só os 4 últimos dígitos), motoristas dos últimos 30 dias e envios dos últimos 14 |
| `GET /api/v1/clientes/:id/whatsapp/previa?dia=AAAA-MM-DD` | colaborador, admin | Mensagens do dia (sem `dia`: ontem), sem enviar |
| `PUT /api/v1/clientes/:id/whatsapp/contatos` `{motorista, telefone, autorizado}` | colaborador, admin | Cadastra ou troca o WhatsApp (exige `autorizado: true`) |
| `PUT /api/v1/clientes/:id/whatsapp/contatos/ativo` `{motorista, ativo}` | colaborador, admin | Pausa ou retoma |
| `DELETE /api/v1/clientes/:id/whatsapp/contatos?motorista=Nome` | colaborador, admin | Apaga o telefone |
| `POST /api/v1/clientes/:id/whatsapp/enviar` `{dia}` | admin | Envia agora (quem já recebeu o dia não recebe de novo) |

A documentação da API Gobrax (externa) está em [../docs/API.md](../docs/API.md).

## WhatsApp dos motoristas

Toda manhã (`WHATSAPP_HORA`, horário de Brasília) cada motorista com WhatsApp cadastrado e ativo recebe a análise individual do dia anterior:

- **Km rodados** e **média de consumo** (km/l = soma dos km ÷ soma dos litros).
- **Início da faixa verde, embalo, motor ligado parado, piloto automático e pedal crítico**, cada um com "melhorou", "piorou" ou "estável" contra a média do próprio motorista nos 7 dias anteriores. O sentido segue o painel: motor ligado parado e pedal crítico pioram quando sobem. Fica estável quem variou menos de 1 ponto (0,05 km/l na média de consumo).
- **Nota do mês**: a Nota Gobrax oficial da premiação, comparada com o mês anterior ("parcial" enquanto o mês não fecha). O nome é casado sem diferenciar maiúsculas nem espaços.
- **Evolução**: "evoluiu", "regrediu um pouco" ou "manteve o nível", pela contagem de indicadores que melhoraram e pioraram.
- **Dica do dia**: o percentual que mais piorou contra a média (piora mínima de 2 pontos), com as mesmas dicas do Diagnóstico do painel.

- O motorista é o nome como vem da API Gobrax. Com mais de um veículo no dia, os valores são juntados pelo tempo dirigido, como no painel.
- O telefone fica cifrado com `CHAVE_TOKENS`; a tela e o log mostram só os 4 últimos dígitos. Só recebe quem tem a autorização confirmada no cadastro.
- No máximo uma mensagem por motorista e dia de resultado (`envios_whatsapp`). Uma mensagem com erro pode ser reenviada pelo botão "Enviar agora".
- Para o motorista parar de receber, o gestor pausa ou remove o telefone na aba WhatsApp.

### Configurar a Twilio (uma vez)

1. **Número WhatsApp Business** ligado à conta Twilio (Messaging > Senders > WhatsApp senders). Ele vai em `WHATSAPP_REMETENTE` no formato `whatsapp:+55DDNNNNNNNNN`.
2. **Template** em Messaging > Content Template Builder: tipo *Text*, idioma português (pt_BR), categoria *Utility*. O texto precisa ser igual, letra por letra, à constante `TextoTemplate` em [modules/whatsapp/resumo.go](modules/whatsapp/resumo.go), que é também o que a aba WhatsApp mostra na prévia. Envie para aprovação do WhatsApp e, quando aprovado, copie o Content SID (`HX...`) para `WHATSAPP_TEMPLATE_SID`. Mudou o texto no código? Crie um template novo e troque o SID.
3. **API Key própria do bot** em Account > API keys & tokens (tipo *Standard*): o SID (`SK...`) vai em `TWILIO_API_KEY` e o segredo em `TWILIO_API_SECRET`. O Account SID (`AC...`) vai em `TWILIO_ACCOUNT_SID`. Não use o Auth Token da conta.
4. Coloque tudo no `.env` (ou nas variáveis da hospedagem), nunca no código, em chat ou em commit. Reinicie o servidor.
5. Cadastre um telefone de teste na aba WhatsApp, confira a prévia e use "Enviar agora" (administrador). Deu certo? Ligue `WHATSAPP_AUTOMATICO=true`.

Variáveis do template: `{{1}}` primeiro nome, `{{2}}` dia (DD/MM), `{{3}}` km, `{{4}}` média km/l, `{{5}}` início da faixa verde, `{{6}}` embalo, `{{7}}` motor ligado parado, `{{8}}` piloto automático, `{{9}}` pedal crítico, `{{10}}` nota do mês, `{{11}}` evolução, `{{12}}` dica. Nenhuma tem quebra de linha (regra do WhatsApp). Na aprovação, informe exemplos realistas para cada uma (por exemplo `{{5}}` = `82,4% ✅ melhorou 3,1 pts`).

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
| WhatsApp não sai | Aba WhatsApp: "Modo de teste" = falta variável da Twilio. "API Key ou segredo recusados" = chave revogada ou copiada errada. Erro 63016 = template não aprovado ou SID errado |

## Limitações conhecidas

- A trava de tentativas fica na memória: zera ao reiniciar e não é compartilhada entre réplicas (rode uma réplica só).
- A sessão tem duração fixa (`SESSAO_HORAS`), sem renovação automática enquanto a pessoa usa.
- Não há registro de auditoria em tabela: só no log (`usuarios: X alterou Y`).
- As telas da plataforma mostram clientes e usuários, mas ainda não os gráficos e indicadores (estes continuam no painel `index.html`).
- O painel em `/painel` usa scripts dentro do HTML, então a CSP dele libera `'unsafe-inline'`; as demais telas continuam com a CSP estrita.
- Apagar no painel um dado que veio da API (não importado à mão) não tem efeito: ele volta na próxima atualização.
- A imagem é construída a partir da raiz do repositório (`docker build -f servidor/Dockerfile .`); no Railway, a pasta raiz do serviço precisa ser a raiz do repositório.
- WhatsApp: o motorista não consegue sair respondendo à mensagem (não há webhook de respostas ainda); quem pausa é o gestor. Situação "enviada" quer dizer aceita pela Twilio, não entregue no celular. Dois motoristas com o mesmo nome no mesmo cliente dividem o cadastro. Um envio interrompido no meio fica "enviando" e não é refeito sozinho.
- A fonte Inter vem do Google Fonts (liberado na CSP); sem internet, cai para a fonte do sistema.

## Próximas fases

2. Telas de indicadores lendo do servidor (substituir o painel no Artifact).
3. HubSpot no servidor.
4. Telas em React (padrão Gobrax).
