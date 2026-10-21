# API Gobrax no painel Análise de Frota

Como o painel busca os dados direto da plataforma Gobrax, sem planilha. O mesmo conteúdo aparece na aba **API Gobrax** do painel (`#/api`), junto com a situação da busca de cada cliente.

O código fica em `index.html`, na seção `API Gobrax` (funções `apiGet`, `diaParaCSV`, `buscarApi` e `buscarAuto`).

## Visão geral

| Item | Valor |
|---|---|
| Base | `https://gateway-v3.gobrax.com.br:8889/api/` |
| Método | `GET` em todos os endpoints |
| Autenticação | `Authorization: Bearer <token>`, um token por cliente |
| Fuso | O dia é o de Brasília (UTC−3). A consulta usa UTC: `AAAA-MM-DD 03:00:00` até `AAAA-MM-DD+1 02:59:59` |
| Limites | Rajada de 10 chamadas, depois 1 por segundo; no máximo 31 dias por consulta |

## Token

- Cada cliente tem o próprio token, digitado no painel em um destes lugares: o perfil do cliente, a lista de clientes em Configurações ou o formulário de novo cliente.
- Ao salvar, o painel tira aspas, espaços, quebras de linha e um `Bearer ` ou `Authorization:` colado junto.
- O token fica só no `localStorage` de quem o digitou, na chave `gbx-api-token:<cliente>`. Ele nunca vai para a base compartilhada, para o código, para o HTML renderizado nem para o repositório.
- Nunca cole um token em chat, e-mail, ticket ou commit. Se isso acontecer, considere o token exposto e peça um novo ao suporte da Gobrax.

## Quando a busca roda

- **Automática:** assim que o token é salvo, o painel busca os últimos 30 dias de estatísticas, performance, odômetro e premiação. Com o painel aberto, a busca se repete a cada hora. Ela começa no último dia buscado, que é buscado de novo porque pode ter vindo parcial. A marca de controle fica em `gbx-api-auto:<cliente>` como `{ em, dia }`.
- **Manual:** no perfil do cliente, o card "Buscar dados pela API Gobrax" busca até 93 dias por vez.
- Todas as buscas entram numa fila e rodam uma de cada vez. O painel espera 120 ms entre chamadas.

## Endpoints

### `GET v1/vehicle-statistics`: estatísticas por veículo

O painel faz uma chamada por dia do período. Cada veículo vira uma linha da planilha de veículos do dia. Os veículos que rodaram (km > 0) entram também na planilha de motoristas, que alimenta a Visão geral.

| Parâmetro | Valor enviado |
|---|---|
| `startDate` | `AAAA-MM-DD 03:00:00` (início do dia de Brasília em UTC) |
| `endDate` | `AAAA-MM-DD+1 02:59:59` |
| `vehicleIdentification` | vazio = todos os veículos do token |

| Campo da API | Coluna no painel |
|---|---|
| `vehicleIdentification` | `veiculo` (placa) |
| `totalMileage` | `km` |
| `totalConsumption` | `litros` |
| `totalBreaking` | `frenagens` |
| `totalBreakingOnHighSpeed` | `frenagens_alta` |
| `averageSpeed` | `velocidade_media` |
| `odometer` | `odometro` |

Os dados aparecem em Veículos e grupos, Veículo individual, Visão geral e Desenvolvimento da frota.

### `GET v1/vehicle-performance`: performance por veículo

Usa a mesma janela diária das estatísticas e é juntada a elas pela placa. Se o token não tiver acesso ou a chamada falhar, as estatísticas do dia entram assim mesmo e o perfil do cliente mostra um aviso.

| Campo da API | Coluna no painel |
|---|---|
| `indicators.greenRange.percentage` | `faixa_verde` |
| `indicators.leverage.percentage` | `embalo` |
| `indicators.idle.percentage` | `motor_ocioso` |
| `indicators.pedalPressureOnHigh.percentage` | `pressao_acelerador` |
| `indicators.powerRange.percentage` | `excesso_rotacao` |
| `indicators.speeding.percentage` | `tempo_excesso_velocidade` |
| `indicators.cruiseControl.percentage` | `piloto_automatico` |
| `indicators.engineBrake.percentage` | `freio_motor` |
| `indicators.movement.duration` | `horas` (minutos ÷ 60) |
| `score` | `nota` (só entre 0 e 100) |
| `drivers[].driverName` | `motorista` (o CPF é descartado) |

### `GET v2/vehicle-odometer`: odômetro

Busca o último odômetro de cada veículo nos últimos 30 dias. O painel guarda um documento por cliente (`alvo: 'odometro'`) e o substitui a cada busca.

| Parâmetro | Valor enviado |
|---|---|
| `startDate` / `endDate` | agora − 30 dias até agora, em UTC (`AAAA-MM-DD HH:MM:SS`) |
| `vehicleIdentification` | vazio = todos |

Os campos usados são `vehicleIdentification`, `odometer` e `lastUpdated`. Eles aparecem em Veículos e grupos, no card Odômetro e comunicação.

### `GET v2/driversOverview`: scores e premiação

Traz a nota, o km e a bonificação de cada motorista por mês. O painel guarda um documento por cliente e mês (`alvo: 'premiacao'`).

| Parâmetro | Valor enviado |
|---|---|
| `startDate` / `endDate` | mês no formato `MM-AAAA` |
| `documentNumbers` | vazio = todos os motoristas |

| Campo da API | No painel |
|---|---|
| `Name` | nome |
| `Overview[].Date` | mês (`MM-AAAA`) |
| `Overview[].Score` | nota |
| `Overview[].TotalKM` | km |
| `Overview[].Reward` | bonificação |
| `DocumentNumber`, `Id` | descartados antes de gravar |

Os dados aparecem na tela Premiação.

## Respostas e erros

| Resposta | Significado | O que o painel faz |
|---|---|---|
| 200 | Sucesso | Lê a lista em `records`, em `data` ou a própria resposta |
| 401 · código 61 | Token inválido | Para a busca e mostra o aviso no perfil |
| 403 · código 71 | Acesso não permitido para o cliente | Na performance, segue só com as estatísticas |
| 429 · código 76 | Limite de chamadas | Espera o `Retry-After` e tenta de novo, até 8 vezes |
| Outros | Erro da API | Mostra o status e a descrição devolvida |

## Onde os dados ficam

- **Painel no Claude:** os dados ficam na base compartilhada, coleção `planilhas`, com um documento por tipo, cliente e dia (ou mês, na premiação).
- **GitHub Pages:** os dados ficam só no navegador (`localStorage`).
- O CPF e o id dos motoristas são descartados antes de qualquer gravação. Só o nome segue.
