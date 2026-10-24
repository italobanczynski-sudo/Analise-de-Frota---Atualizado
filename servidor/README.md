# Servidor do Análise de Frota (Go)

Busca os dados da API Gobrax de cada cliente, guarda no Postgres e entrega pela API HTTP. Substitui o `api/puxar.mjs` e, a partir da fase 2, vai alimentar as telas do painel. Segue o padrão Go da Gobrax (Clean Architecture: `adapters` → `modules` → `api` → `background`).

## Rodar no computador (Docker)

Precisa só do Docker Desktop.

1. Copie `.env.example` para `.env` e preencha (o `.env` fica fora do Git):
   - `POSTGRES_PASSWORD`: senha do banco local.
   - `CHAVE_TOKENS`: chave que cifra os tokens no banco, gerada com `openssl rand -base64 32`. Se ela mudar, os tokens precisam ser salvos de novo.
   - `CHAVE_ADMIN`: chave de acesso à API (cabeçalho `X-Chave-Admin`), com pelo menos 24 caracteres. Vale até a fase de login.
2. Suba o banco e o servidor:
   ```bash
   docker compose up -d --build
   ```
3. Cadastre os clientes. Para trazer os do script antigo (`api/clientes.json` e os tokens do `api/.env`):
   ```bash
   docker compose run --rm app -importar-clientes /importar/clientes.json -tokens /importar/.env
   ```
   Ou pela API: `POST /api/v1/clientes` com `{"nome", "hubspot_id", "token"}`.

Testes (dentro de um contêiner Go, sem instalar nada):
```bash
docker compose --profile testes run --rm testes
```

Parar: `docker compose down` (os dados ficam no volume `dados`; `docker compose down -v` apaga o banco).

## O que ele faz sozinho

Logo que sobe e depois a cada `COLETA_INTERVALO_MIN` (60 min), busca todos os clientes com token, um de cada vez:

- **Dia a dia** (`v1/vehicle-statistics` + `v1/vehicle-performance`): uma chamada de cada por dia de Brasília. Na primeira vez, os últimos `COLETA_DIAS` (30); depois, do último dia buscado (que é refeito, pode ter vindo parcial) até hoje. As regras de junção são as do painel: motorista do dia é quem dirigiu por mais tempo, veículo parado entra com 0 km e sem indicadores, a nota só vale entre 0 e 100.
- **Odômetro** (`v2/vehicle-odometer`): último de cada veículo nos últimos 30 dias.
- **Premiação** (`v2/driversOverview`): do mês do início até o mês atual.

Falhas: 404 é dia sem dado. Uma falha pontual pula o dia sem apagar o que já estava salvo, e a próxima busca tenta de novo. Três falhas seguidas desligam aquele tipo na rodada. Token inválido para a busca do cliente e fica registrado na situação.

## Segurança

- O token de cada cliente é cifrado (AES-256-GCM) antes de ir para o banco e nunca sai em nenhuma resposta da API; a lista de clientes só diz `tem_token`.
- O CPF que a API devolve nunca é lido: os structs do adapter nem declaram esse campo. Só o nome do motorista é guardado.
- Toda rota em `/api/v1` exige o cabeçalho `X-Chave-Admin`.
- Nunca cole token, chave ou o conteúdo do `.env` em chat, e-mail ou commit.

## API

Respostas: `{"ok": true, "dados": ...}` ou `{"ok": false, "erro": "..."}`.

| Método e rota | O que faz |
|---|---|
| `GET /saude` | Responde se o servidor está no ar (sem chave) |
| `GET /api/v1/clientes` | Clientes, com a situação da última busca |
| `POST /api/v1/clientes` | Cadastra `{"nome", "hubspot_id", "token"}` (token opcional) |
| `PUT /api/v1/clientes/:id/token` | Troca o token `{"token"}` |
| `POST /api/v1/clientes/:id/coleta` | Busca agora, em segundo plano (202); 409 se já estiver buscando |
| `GET /api/v1/clientes/:id/coleta` | Situação da busca: `nunca`, `rodando`, `ok` ou `erro`, último dia completo e avisos |
| `GET /api/v1/clientes/:id/registros?inicio=AAAA-MM-DD&fim=AAAA-MM-DD` | Dia a dia de cada veículo |
| `GET /api/v1/clientes/:id/odometros` | Último odômetro de cada veículo |
| `GET /api/v1/clientes/:id/premiacoes?mes=AAAA-MM` | Premiação de um mês (sem `mes`: todos) |

## Estrutura

```
environment/   única leitura de variáveis de ambiente (Config_t carregado uma vez)
adapters/      Postgres (Singleton), migrações SQL, cliente da API Gobrax, log
common/        datas de Brasília, placas, limpeza de token, validação, criptografia
modules/       cliente, diario, odometro, premiacao, coleta (regras de negócio + repositórios)
api/           controller (entrada HTTP) e middleware (casos de uso), rotas, chave de acesso
background/    busca automática (gocron)
```

## Próximas fases

2. Endpoints para cada tela e o painel lendo do servidor.
3. Login e permissões; HubSpot no servidor.
4. Telas em React (padrão Gobrax).
