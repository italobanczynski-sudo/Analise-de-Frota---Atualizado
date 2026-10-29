#!/bin/sh
# Teste ponta a ponta do login, das sessões e das permissões, contra o servidor no ar.
# CRIA USUÁRIOS E CLIENTES: rode só num banco de teste (testes/rodar-e2e.sh já monta um descartável).
# Variáveis: BASE (endereço), CONVITE_ADMIN (token do convite do primeiro admin);
# opcionais: LOGS_CMD (lê o log do servidor, para o código por e-mail) e SQL_CMD (psql no banco de teste).
set -u
BASE=${BASE:-http://localhost:8080}
TMP=$(mktemp -d)
falhas=0
total=0

confere() { # descrição esperado obtido
  total=$((total + 1))
  if [ "$2" = "$3" ]; then echo "ok    $1"; else echo "FALHA $1 (esperado $2, veio $3)"; falhas=$((falhas + 1)); fi
}
contem() { # descrição texto (no corpo da última resposta)
  total=$((total + 1))
  if grep -q "$2" "$TMP/resp"; then echo "ok    $1"; else echo "FALHA $1 (resposta sem \"$2\": $(head -c 200 "$TMP/resp"))"; falhas=$((falhas + 1)); fi
}
nao_contem() {
  total=$((total + 1))
  if grep -q "$2" "$TMP/resp"; then echo "FALHA $1 (resposta tem \"$2\")"; falhas=$((falhas + 1)); else echo "ok    $1"; fi
}
# req quem método caminho [corpo] [ip]: imprime o status; o corpo fica em $TMP/resp e o cookie em $TMP/<quem>
req() {
  set -- "$1" "$2" "$3" "${4:-}" "${5:-198.51.100.1}"
  if [ -n "$4" ]; then
    curl -s -o "$TMP/resp" -w '%{http_code}' -b "$TMP/$1" -c "$TMP/$1" -H 'X-Requisicao: plataforma' -H "X-Forwarded-For: $5" \
      -H 'Content-Type: application/json' -X "$2" "$BASE$3" -d "$4"
  else
    curl -s -o "$TMP/resp" -w '%{http_code}' -b "$TMP/$1" -c "$TMP/$1" -H 'X-Requisicao: plataforma' -H "X-Forwarded-For: $5" -X "$2" "$BASE$3"
  fi
}
campo() { sed -n "s/.*\"$1\":\"\{0,1\}\([^\",}]*\).*/\1/p" "$TMP/resp" | head -1; } # valor simples do JSON
token_do_link() { sed -n 's/.*"link_convite":"[^"#]*#\([A-Za-z0-9_-]*\)".*/\1/p' "$TMP/resp"; }
cookie() { awk '$6 == "rg_sessao" { print $7 }' "$TMP/$1"; }
SENHA='SenhaE2e12345'

echo "== páginas e cabeçalhos"
confere "tela de entrada abre" 200 "$(curl -s -o "$TMP/resp" -w '%{http_code}' "$BASE/")"
confere "saúde" 200 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/saude")"
csp=$(curl -s -D - -o /dev/null "$BASE/" | grep -ci "content-security-policy: default-src 'self'")
confere "CSP presente" 1 "$csp"
confere "pasta /estatico/ não é listada" 404 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/estatico/")"
confere "CSS da tela de entrada" 200 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/estatico/login.css")"
acao=$(curl -s -D - -o /dev/null -H 'Origin: https://site-malicioso.exemplo' "$BASE/saude" | grep -ci "access-control-allow-origin")
confere "CORS recusa outra origem" 0 "$acao"

echo "== sem login"
confere "lista de clientes sem login" 401 "$(req anon GET /api/v1/clientes)"
confere "lista de usuários sem login" 401 "$(req anon GET /api/v1/usuarios)"
confere "login sem o cabeçalho da plataforma (CSRF)" 403 \
  "$(curl -s -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' -X POST "$BASE/api/v1/auth/entrar" -d '{"email":"a@b.com","senha":"x"}')"
confere "cookie inventado" 401 "$(curl -s -o /dev/null -w '%{http_code}' -b 'rg_sessao=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA' "$BASE/api/v1/auth/eu")"

echo "== convite do administrador"
confere "convite com token errado" 400 "$(req x POST /api/v1/auth/convite '{"token":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","senha":"'$SENHA'"}')"
confere "convite com senha fraca" 400 "$(req admin POST /api/v1/auth/convite '{"token":"'"$CONVITE_ADMIN"'","senha":"curta"}')"
confere "admin aceita o convite" 200 "$(req admin POST /api/v1/auth/convite '{"token":"'"$CONVITE_ADMIN"'","senha":"'$SENHA'"}')"
nao_contem "resposta não traz hash de senha" "senha_hash"
confere "convite não vale duas vezes" 400 "$(req outro POST /api/v1/auth/convite '{"token":"'"$CONVITE_ADMIN"'","senha":"'$SENHA'"}')"
confere "admin logado" 200 "$(req admin GET /api/v1/auth/eu)"
contem "perfil admin" '"papel":"admin"'

echo "== cadastro de pessoas (admin)"
confere "convida colaborador" 201 "$(req admin POST /api/v1/usuarios '{"nome":"Colab E2E","email":"colab.e2e@exemplo.com","papel":"colaborador"}')"
TK_COLAB=$(token_do_link); ID_COLAB=$(sed -n 's/.*"usuario":{"id":\([0-9]*\).*/\1/p' "$TMP/resp")
confere "convida leitor" 201 "$(req admin POST /api/v1/usuarios '{"nome":"Leitor E2E","email":"leitor.e2e@exemplo.com","papel":"leitor"}')"
TK_LEITOR=$(token_do_link); ID_LEITOR=$(sed -n 's/.*"usuario":{"id":\([0-9]*\).*/\1/p' "$TMP/resp")
confere "e-mail repetido" 409 "$(req admin POST /api/v1/usuarios '{"nome":"Outro","email":"COLAB.e2e@exemplo.com","papel":"leitor"}')"
confere "perfil inexistente" 400 "$(req admin POST /api/v1/usuarios '{"nome":"Outro","email":"x.e2e@exemplo.com","papel":"dono"}')"
confere "e-mail inválido" 400 "$(req admin POST /api/v1/usuarios '{"nome":"Outro","email":"sem-arroba","papel":"leitor"}')"
confere "colaborador cria a senha" 200 "$(req colab POST /api/v1/auth/convite '{"token":"'"$TK_COLAB"'","senha":"'$SENHA'"}')"
confere "convida pessoa para o teste da trava" 201 "$(req admin POST /api/v1/usuarios '{"nome":"Trava E2E","email":"trava.e2e@exemplo.com","papel":"leitor"}')"
TK_TRAVA=$(token_do_link)
confere "pessoa da trava cria a senha" 200 "$(req trava POST /api/v1/auth/convite '{"token":"'"$TK_TRAVA"'","senha":"'$SENHA'"}')"
confere "leitor cria a senha" 200 "$(req leitor POST /api/v1/auth/convite '{"token":"'"$TK_LEITOR"'","senha":"'$SENHA'"}')"

echo "== clientes (admin)"
confere "admin cadastra cliente com token" 201 "$(req admin POST /api/v1/clientes '{"nome":"Cliente E2E","token":"token.de.teste-123"}')"
ID_CLI=$(campo id)
contem "resposta diz tem_token" '"tem_token":true'
nao_contem "token nunca volta na resposta" "token.de.teste"
confere "cliente sem nome" 400 "$(req admin POST /api/v1/clientes '{"nome":""}')"
confere "cliente repetido" 409 "$(req admin POST /api/v1/clientes '{"nome":"Cliente E2E"}')"
confere "HubSpot com letras" 400 "$(req admin POST /api/v1/clientes '{"nome":"Outro","hubspot_id":"abc"}')"
confere "admin cadastra cliente sem token" 201 "$(req admin POST /api/v1/clientes '{"nome":"Sem Token E2E"}')"
ID_SEM=$(campo id)
confere "busca de cliente sem token" 400 "$(req admin POST /api/v1/clientes/$ID_SEM/coleta)"
confere "lista de clientes" 200 "$(req admin GET /api/v1/clientes)"
nao_contem "lista sem token" "token.de.teste"
confere "registros com datas válidas" 200 "$(req admin GET "/api/v1/clientes/$ID_CLI/registros?inicio=2026-10-01&fim=2026-10-02")"
confere "registros sem datas" 400 "$(req admin GET /api/v1/clientes/$ID_CLI/registros)"
confere "registros com período invertido" 400 "$(req admin GET "/api/v1/clientes/$ID_CLI/registros?inicio=2026-10-05&fim=2026-10-01")"
confere "id de cliente inválido" 400 "$(req admin GET /api/v1/clientes/abc/odometros)"
confere "cliente inexistente" 404 "$(req admin GET /api/v1/clientes/999999/odometros)"
confere "premiação com mês inválido" 400 "$(req admin GET "/api/v1/clientes/$ID_CLI/premiacoes?mes=2026-13")"
confere "premiação" 200 "$(req admin GET "/api/v1/clientes/$ID_CLI/premiacoes?mes=2026-10")"
confere "situação da busca" 200 "$(req admin GET /api/v1/clientes/$ID_CLI/coleta)"
confere "mensagens enviadas do cliente" 200 "$(req admin GET /api/v1/clientes/$ID_CLI/mensagens/enviadas)"
contem "sem mensagens enviadas ainda" '"dados":\[\]'
confere "mensagens enviadas de cliente inexistente" 404 "$(req admin GET /api/v1/clientes/999999/mensagens/enviadas)"

echo "== permissões do leitor"
confere "leitor vê clientes" 200 "$(req leitor GET /api/v1/clientes)"
confere "leitor vê registros" 200 "$(req leitor GET "/api/v1/clientes/$ID_CLI/registros?inicio=2026-10-01&fim=2026-10-02")"
confere "leitor não pede busca" 403 "$(req leitor POST /api/v1/clientes/$ID_CLI/coleta)"
confere "leitor não cadastra cliente" 403 "$(req leitor POST /api/v1/clientes '{"nome":"Invasor"}')"
confere "leitor não troca token" 403 "$(req leitor PUT /api/v1/clientes/$ID_CLI/token '{"token":"abc"}')"
confere "leitor não lista usuários" 403 "$(req leitor GET /api/v1/usuarios)"
confere "leitor não cria usuário" 403 "$(req leitor POST /api/v1/usuarios '{"nome":"X","email":"x2.e2e@exemplo.com","papel":"admin"}')"
confere "leitor não vira admin" 403 "$(req leitor PUT /api/v1/usuarios/$ID_LEITOR '{"papel":"admin","ativo":true}')"
confere "leitor não gera convite" 403 "$(req leitor POST /api/v1/usuarios/$ID_COLAB/convite)"

echo "== painel e base do painel"
confere "painel sem login volta para a entrada" 302 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/painel")"
confere "painel com login" 200 "$(req admin GET /painel)"
contem "painel ligado à base do servidor" "/estatico/painel-base.js"
contem "painel com o ícone da Gobrax na aba" 'rel="icon" href="/estatico/icone.svg"'
confere "favicon.ico vai para o ícone" 301 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/favicon.ico")"
confere "dados da API para o painel" 200 "$(req leitor GET /api/v1/painel/dados)"
contem "clientes nos dados do painel" "Cliente E2E"
confere "admin grava documento" 200 "$(req admin PUT /api/v1/base/tratativas/t1 '{"cliente":"Cliente E2E","etapa":"aberta"}')"
confere "leitor lê a base" 200 "$(req leitor GET /api/v1/base/tratativas)"
contem "documento gravado volta" '"etapa":"aberta"'
confere "leitor não grava na base" 403 "$(req leitor PUT /api/v1/base/tratativas/t2 '{"x":1}')"
confere "leitor não apaga da base" 403 "$(req leitor DELETE /api/v1/base/tratativas/t1)"
confere "coleção fora da lista" 400 "$(req admin PUT /api/v1/base/usuarios/x '{"x":1}')"
confere "documento que não é objeto" 400 "$(req admin PUT /api/v1/base/tratativas/t3 '[1,2]')"
confere "colaborador cria documento" 201 "$(req colab POST /api/v1/base/tratativas '{"cliente":"Cliente E2E"}')"
confere "colaborador apaga documento" 200 "$(req colab DELETE /api/v1/base/tratativas/t1)"

echo "== permissões do colaborador"
confere "colaborador não lista usuários" 403 "$(req colab GET /api/v1/usuarios)"
confere "colaborador não cadastra cliente" 403 "$(req colab POST /api/v1/clientes '{"nome":"Invasor"}')"
confere "colaborador não troca token" 403 "$(req colab PUT /api/v1/clientes/$ID_CLI/token '{"token":"abc"}')"
confere "colaborador passa do perfil na busca (cliente inexistente)" 404 "$(req colab POST /api/v1/clientes/999999/coleta)"

echo "== senha errada e trava de tentativas"
confere "senha errada" 401 "$(req tmp POST /api/v1/auth/entrar '{"email":"trava.e2e@exemplo.com","senha":"errada123456"}' 203.0.113.1)"
contem "mensagem genérica" "e-mail ou senha incorretos"
confere "e-mail desconhecido dá a mesma resposta" 401 "$(req tmp POST /api/v1/auth/entrar '{"email":"ninguem.e2e@exemplo.com","senha":"errada123456"}' 203.0.113.1)"
contem "mesma mensagem" "e-mail ou senha incorretos"
for i in 1 2 3 4; do req tmp POST /api/v1/auth/entrar '{"email":"trava.e2e@exemplo.com","senha":"errada123456"}' "203.0.113.1$i" >/dev/null; done
confere "5 erros travam o e-mail, mesmo com a senha certa" 401 "$(req tmp POST /api/v1/auth/entrar '{"email":"trava.e2e@exemplo.com","senha":"'$SENHA'"}' 203.0.113.50)"
contem "mensagem da trava" "muitas tentativas"
for i in $(seq 1 20); do req tmp POST /api/v1/auth/entrar '{"email":"robo'$i'.e2e@exemplo.com","senha":"errada123456"}' 203.0.113.99 >/dev/null; done
confere "20 erros travam o IP" 401 "$(req tmp POST /api/v1/auth/entrar '{"email":"colab.e2e@exemplo.com","senha":"'$SENHA'"}' 203.0.113.99)"
confere "outra pessoa, de outro IP, entra normalmente" 200 "$(req colab2 POST /api/v1/auth/entrar '{"email":"colab.e2e@exemplo.com","senha":"'$SENHA'"}' 203.0.113.100)"

echo "== sair e sessão"
COOKIE_VELHO=$(cookie colab2)
confere "sair" 200 "$(req colab2 POST /api/v1/auth/sair)"
confere "depois de sair, sem acesso" 401 "$(req colab2 GET /api/v1/auth/eu)"
confere "o cookie antigo não vale mais" 401 "$(curl -s -o /dev/null -w '%{http_code}' -b "rg_sessao=$COOKIE_VELHO" "$BASE/api/v1/auth/eu")"

if [ -n "${SQL_CMD:-}" ]; then
  confere "colaborador entra de novo" 200 "$(req colab3 POST /api/v1/auth/entrar '{"email":"colab.e2e@exemplo.com","senha":"'$SENHA'"}' 203.0.113.101)"
  $SQL_CMD "UPDATE sessoes SET expira_em = now() - interval '1 minute' WHERE usuario_id = $ID_COLAB" >/dev/null
  confere "sessão vencida não entra" 401 "$(req colab3 GET /api/v1/auth/eu)"
fi

echo "== troca de perfil e desativação (admin)"
confere "leitor logado antes da troca" 200 "$(req leitor GET /api/v1/auth/eu)"
confere "admin promove leitor a colaborador" 200 "$(req admin PUT /api/v1/usuarios/$ID_LEITOR '{"papel":"colaborador","ativo":true}')"
confere "a sessão antiga do leitor caiu" 401 "$(req leitor GET /api/v1/auth/eu)"
confere "admin desativa colaborador" 200 "$(req admin PUT /api/v1/usuarios/$ID_COLAB '{"papel":"colaborador","ativo":false}')"
confere "desativado não entra" 403 "$(req tmp POST /api/v1/auth/entrar '{"email":"colab.e2e@exemplo.com","senha":"'$SENHA'"}' 203.0.113.102)"
confere "não gera convite para desativado" 403 "$(req admin POST /api/v1/usuarios/$ID_COLAB/convite)"
ID_ADMIN=$(req admin GET /api/v1/auth/eu >/dev/null; campo id)
confere "único admin não deixa de ser admin" 409 "$(req admin PUT /api/v1/usuarios/$ID_ADMIN '{"papel":"leitor","ativo":true}')"
confere "único admin não se desativa" 409 "$(req admin PUT /api/v1/usuarios/$ID_ADMIN '{"papel":"admin","ativo":false}')"
confere "usuário inexistente" 404 "$(req admin PUT /api/v1/usuarios/999999 '{"papel":"leitor","ativo":true}')"

echo "== esqueci minha senha (código por e-mail)"
confere "pedido para e-mail desconhecido responde igual" 200 "$(req tmp POST /api/v1/auth/codigo '{"email":"ninguem.e2e@exemplo.com"}' 203.0.113.120)"
contem "mensagem neutra" "Se este e-mail estiver cadastrado"
confere "código errado" 400 "$(req tmp POST /api/v1/auth/codigo/confirmar '{"email":"leitor.e2e@exemplo.com","codigo":"000000","senha":"'$SENHA'X1"}' 203.0.113.121)"
if [ -n "${LOGS_CMD:-}" ]; then
  confere "leitor entra (sessão que deve cair)" 200 "$(req leitor2 POST /api/v1/auth/entrar '{"email":"leitor.e2e@exemplo.com","senha":"'$SENHA'"}' 203.0.113.122)"
  confere "leitor pede código" 200 "$(req tmp POST /api/v1/auth/codigo '{"email":"leitor.e2e@exemplo.com"}' 203.0.113.123)"
  confere "segundo pedido antes de 1 minuto" 429 "$(req tmp POST /api/v1/auth/codigo '{"email":"leitor.e2e@exemplo.com"}' 203.0.113.123)"
  sleep 1
  CODIGO=$($LOGS_CMD 2>/dev/null | grep "para leitor.e2e@exemplo.com: Seu código de acesso" | tail -1 | sed -n 's/.*acesso: \([0-9]\{6\}\).*/\1/p')
  confere "código achado no log de teste" 6 "${#CODIGO}"
  NOVA='NovaSenhaE2e987'
  confere "código certo grava a senha nova" 200 "$(req leitor3 POST /api/v1/auth/codigo/confirmar '{"email":"leitor.e2e@exemplo.com","codigo":"'"$CODIGO"'","senha":"'$NOVA'"}' 203.0.113.124)"
  confere "já entra logado" 200 "$(req leitor3 GET /api/v1/auth/eu)"
  confere "a sessão aberta com a senha velha caiu" 401 "$(req leitor2 GET /api/v1/auth/eu)"
  confere "código não vale duas vezes" 400 "$(req tmp POST /api/v1/auth/codigo/confirmar '{"email":"leitor.e2e@exemplo.com","codigo":"'"$CODIGO"'","senha":"OutraSenha1234"}' 203.0.113.125)"
  confere "senha velha não entra" 401 "$(req tmp POST /api/v1/auth/entrar '{"email":"leitor.e2e@exemplo.com","senha":"'$SENHA'"}' 203.0.113.126)"
  confere "senha nova entra" 200 "$(req tmp POST /api/v1/auth/entrar '{"email":"leitor.e2e@exemplo.com","senha":"'$NOVA'"}' 203.0.113.127)"
  contem "perfil novo (colaborador)" '"papel":"colaborador"'
fi

rm -rf "$TMP"
echo
echo "$((total - falhas)) de $total verificações passaram"
[ "$falhas" -eq 0 ]
