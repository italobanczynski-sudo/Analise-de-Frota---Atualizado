# API Gobrax no painel Análise de Frota

Como o painel busca os dados direto da plataforma Gobrax, sem planilha. O mesmo conteúdo aparece na aba **API Gobrax** do painel (`#/api`), junto com a situação da busca de cada cliente.

O código fica em `index.html`, na seção `API Gobrax` (funções `apiGet`, `diaParaCSV`, `buscarApi` e `buscarAuto`). O script `api/puxar.mjs` repete as mesmas regras fora do navegador.

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

## Puxar pelo script do projeto (`api/puxar.mjs`)

O painel no Claude não deixa a página chamar endereços de fora, então a busca pelo navegador falha lá. A API aceita chamadas de qualquer origem, por isso no GitHub Pages ela funciona. Para alimentar o painel no Claude, use o script do projeto. Ele precisa do Node 18 ou mais novo.

1. Copie `api/.env.example` para `api/.env` e cole o token de cada cliente depois do `=`:
   ```
   GOBRAX_TOKEN_XTSUL=<token do cliente>
   ```
2. Copie `api/clientes.example.json` para `api/clientes.json` e escreva o nome de cada cliente como está no painel, junto com o nome da variável do token:
   ```json
   [{ "cliente": "Xtsul Transportes LTDA", "token": "GOBRAX_TOKEN_XTSUL" }]
   ```
3. Rode `node api/puxar.mjs`. Por padrão ele busca os últimos 30 dias. As opções são `--dias 7` e `--cliente "Xtsul Transportes LTDA"`.

Os arquivos `api/.env`, `api/clientes.json` e `api/saida/` ficam fora do Git (`.gitignore`). Nunca faça commit deles e nunca cole o token em chat.

O script gera em `api/saida/`:
- `<cliente>/*.json`: um arquivo por documento, no formato da coleção `planilhas` da base compartilhada (motoristas e veículos por dia, odômetro, premiação por mês). Para carregar no painel do Claude, peça ao Claude Code: "carrega api/saida no painel". Ele grava os arquivos na base, e quem estiver com o painel aberto vê os dados na hora.
- `<cliente>-veiculos.csv` e `<cliente>-motoristas.csv`: planilhas para importar no perfil do cliente, em qualquer versão do painel.

Para rodar sozinho a cada hora no Windows, crie uma tarefa agendada:
```
schtasks /Create /SC HOURLY /TN "Gobrax API" /TR "node \"C:\caminho\do\projeto\api\puxar.mjs\" --dias 2"
```
A tarefa atualiza `api/saida/`. Para levar os dados ao painel do Claude, ainda é preciso pedir a carga ao Claude Code.

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
| `indicators.extraEconomicRange.percentage` | `inicio_faixa_verde` |
| `indicators.economicRange.percentage` | `final_faixa_verde` |
| `indicators.ecoRoll.percentage` | `eco_roll` |
| `indicators.pedalPressureOnLow.percentage` | `pressao_ideal` |
| `indicators.pedalPressureOnMid.percentage` | `pressao_atencao` |
| `indicators.movement.duration` | `horas` (minutos ÷ 60) |
| `score` | `nota` (só entre 0 e 100) |
| `drivers[].driverName` | `motorista`: com mais de um no dia, fica quem dirigiu por mais tempo; sem motorista, o dia entra só no veículo (o CPF é descartado) |

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

## O que a API respondeu de verdade (teste com o token da Xtsul, 21/10/2026)

| Endpoint | Resultado |
|---|---|
| `v1/vehicle-performance` | Funciona. Documentação oficial: https://documenter.getpostman.com/view/27100945/2s93eU4FLe |
| `v1/vehicle-statistics` | HTTP 400 com corpo vazio, mesmo com placa e com outros formatos de data. Não está na documentação pública; confirmar com o suporte se o token precisa de liberação |
| `v2/vehicle-odometer` | HTTP 400 com corpo vazio, como as estatísticas. Documentação: https://documenter.getpostman.com/view/27100945/2s93sf2BRK. Testado no formato exato da documentação, com e sem placa (com e sem hífen), espaço codificado ou não e na porta 8888: sempre 400 |
| `v2/driversOverview` | HTTP 500, código 9 ("Erro não esperado"), para qualquer intervalo de meses (testado de 2024 a 2026). Com o mesmo mês no início e no fim, responde 400, código 67 ("Datas fornecidas inválidas"). Documentação: https://documenter.getpostman.com/view/27100945/2s9Ye8hvPt |

Detalhes da resposta da performance:
- Não existe nota geral (`score`) no registro. Cada indicador traz o próprio `score`, por isso a coluna `nota` fica vazia.
- A pressão crítica do pedal vem como `pedalPressureOnHig`, sem o "h" final. O painel e o script aceitam os dois nomes.
- `movement.duration` vem em minutos.
- A pressão crítica do pedal vem como `pedalPressureOnHig` (sem o "h"); o painel aceita os dois nomes.
- Sem nota e sem km/l, o status de cada motorista e veículo (evoluiu, estável, regrediu) usa a faixa verde: ±2 p.p. contra o período anterior.
- Sem km, conta como rodando quem teve tempo em movimento, e as médias usam as horas como peso.
- Modelo da resposta, com os nomes dos campos e sem dados reais: `{"records":[{"vehicleIdentification":"","startDate":"","endDate":"","drivers":[{"driverName":"","cpf":"","startDate":"","endDate":""}],"indicators":{"movement":{"duration":0,"percentage":0,"score":0}, …}}]}`.
- Sem as estatísticas, o script monta as linhas só com a performance: placa, motorista, horas e indicadores, sem km e sem litros. Conta como "rodou" quem teve tempo em movimento.
- Os erros vêm em `{"success": false, "data": {"code", "description"}}`; a documentação da performance cita também `{"errorCode", "errorDescription"}`.

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
