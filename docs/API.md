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
   GOBRAX_TOKEN_XTSUL=<token da Xtsul>
   GOBRAX_TOKEN_CAJU=<token da Caju>
   ```
2. Copie `api/clientes.example.json` para `api/clientes.json` e escreva o nome de cada cliente como está no painel, junto com o nome da variável do token e o ID da empresa no HubSpot (o número no fim do link da empresa):
   ```json
   [
     { "cliente": "Xtsul Transportes LTDA", "token": "GOBRAX_TOKEN_XTSUL", "hubspot_id": "<ID no HubSpot>" },
     { "cliente": "Caju Transportes", "token": "GOBRAX_TOKEN_CAJU", "hubspot_id": "<ID no HubSpot>" }
   ]
   ```
3. Rode `node api/puxar.mjs`. Por padrão ele busca os últimos 30 dias. As opções são `--dias 7` e `--cliente "Caju Transportes"` (só um cliente). O HubSpot segue o mesmo arquivo: `node api/hubspot.mjs --cliente "Caju Transportes"`.

Para incluir um cliente novo: uma linha `GOBRAX_TOKEN_<NOME>=` no `api/.env` e uma entrada no `api/clientes.json`. O cliente precisa existir no painel com o mesmo nome.

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
| `drivers[].driverName`, `startDate`, `endDate` | `motorista` (o CPF é descartado). O veículo fica com o dia inteiro. O motorista fica só com o tempo em que estava vinculado: um único motorista vinculado o dia todo usa a linha do veículo; nos demais casos o painel consulta estatísticas e performance com `vehicleIdentification` e o início e o fim de cada vínculo, e soma as janelas (km, litros e freadas somados; indicadores pela duração sobre o tempo em movimento). Sem motorista, o dia entra só no veículo |

Motorista não é veículo: na plataforma, a análise por motorista mostra só o tempo em que ele estava no veículo. Antes, o painel dava ao motorista o dia inteiro do veículo (quem dirigiu mais tempo ficava com tudo), e os números do condutor saíam maiores que os da plataforma.

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

## O que a API respondeu de verdade (token da Xtsul, 21/10/2026)

As quatro APIs funcionam. Estatísticas, hodômetro e premiação responderam 400/500 durante parte do dia e voltaram depois; o script tolera falhas pontuais:
- uma falha pula só aquele dia e o aviso lista os dias que faltaram (rode de novo para completar);
- 404 ("Veículo não identificado") é tratado como dia sem dado;
- três falhas seguidas desligam aquele tipo de busca na rodada.

Documentações oficiais: performance https://documenter.getpostman.com/view/27100945/2s93eU4FLe · hodômetro https://documenter.getpostman.com/view/27100945/2s93sf2BRK · premiação https://documenter.getpostman.com/view/27100945/2s9Ye8hvPt

Nota Gobrax: a performance não traz a nota geral, só a nota de cada indicador. Testamos recalcular a nota com os pesos do treinamento (início da faixa verde 45, embalo 25, motor ligado parado 15, acima do verde 10, excesso de velocidade 5, piloto automático 5) e comparamos com a nota oficial mensal da premiação: erro médio de 5 a 6 pontos, até 15. Por isso o painel não estima a nota; a nota oficial aparece na tela Premiação, mês a mês, e a evolução dos motoristas usa a média de km/l.

Premiação (`v2/driversOverview`): o modelo da resposta é `{"success": true, "data": [{"ID", "Name", "DocumentNumber", "Overview": [{"Date", "Reward", "TotalKM", "Score"}]}]}`. O painel usa `Name`, `Date`, `Score`, `TotalKM` e `Reward`; `ID` e `DocumentNumber` (CPF) são descartados. A tela Premiação mostra o mês escolhido e a evolução mês a mês de cada motorista.

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

## Preço do diesel por estado

- Fonte: `GET https://combustivelapi.com.br/api/precos/` (preços da Petrobras, aceita chamada do navegador). O painel usa `precos.diesel` (`br` e a sigla de cada estado, em reais com vírgula).
- No perfil do cliente, o cartão "Onde o cliente opera" guarda o estado do cliente (`uf`) e, se quiser, um estado por grupo de veículos (`ufGrupos`, chave = nome do grupo normalizado).
- Na barra do topo, um clique na pílula verde do preço abre a lista para escolher: "Automático" (o padrão, regra abaixo), "Média Brasil (geral)" ou qualquer estado. A escolha vale só no navegador de quem escolheu (`gbx-diesel-uf`).
- Preço usado nas contas no automático: o do estado do grupo filtrado; senão o do cliente; senão a média Brasil. Estado sem preço na fonte usa a média Brasil. Um preço digitado à mão continua valendo até voltar para o automático.
- No GitHub Pages o painel busca ao abrir e a cada 6 horas. No Claude a página não chama fora: o script `api/puxar.mjs` grava `api/saida/diesel_uf.json`, carregado na base em `config/diesel_uf`.

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

## HubSpot (perfil e relacionamento do cliente)

O cartão **Cliente no HubSpot** e o **Estudo do cliente** (Saúde do cliente) usam duas fontes, a que for mais recente:

1. **Conector do HubSpot no claude.ai** (painel aberto dentro do Claude): lido ao vivo com a autorização de quem abre o painel. Não há token no painel.
2. **API do HubSpot por script** (`api/hubspot.mjs`): para usar fora do Claude ou guardar uma leitura.
   - Crie um app privado no HubSpot (Configurações → Integrações → Apps privados) só com escopos de leitura: empresas, contatos, tickets, e-mails, ligações e reuniões.
   - Coloque o token no `api/.env` como `HUBSPOT_TOKEN=...` (o arquivo fica fora do Git; nunca cole o token em chat).
   - No `api/clientes.json`, `hubspot_id` é o id da empresa no HubSpot (opcional; sem ele, busca pelo nome).
   - Rode `node api/hubspot.mjs` (ou `--cliente "Nome"`). Gera `api/saida/<cliente>/hubspot.json`.
   - No painel, em Saúde do cliente → Cliente no HubSpot → **Importar arquivo**. A leitura fica guardada na base compartilhada (coleção `hubspot`) e vale até o conector trazer uma mais nova.

Endpoints usados (só leitura): `GET /crm/v3/objects/companies/{id}`, `POST /crm/v3/objects/companies/search`, `GET /crm/v4/objects/{tipo}/{id}/associations/{tipo}` e `POST /crm/v3/objects/{tipo}/batch/read`.

Campos da empresa: frota (`frota_contratada`, `frota_total`), nota CHS (`nota_do_chs`, `chs_score` e as partes `chs_aderencia`, `chs_nota_engajamento`, `chs_nota_adimplencia`, `chs_nota_usuario`, `chs_nota_motorista`, `chs_nota_tickets_positivos`), contrato (`data_inicio_do_contrato_original`, `data_de_termino_do_contrato`, `ciclo_experiencia`), `estados_onde_opera` e último contato (`notes_last_contacted`).

Dados pessoais ficam de fora (Política de Segurança da Gobrax, cláusula 5.1): o script e o painel não pedem CNPJ, e-mail nem telefone; dos contatos vêm só nome, cargo e último contato. E-mails, CPF, CNPJ e telefones escritos em notas, reuniões, ligações e assuntos saem mascarados (`[e-mail]`, `[CPF]`, `[CNPJ]`, `[telefone]`). Um arquivo antigo que ainda tenha esses dados é limpo e regravado quando o painel o lê.
