#!/usr/bin/env bash
# Sobe um ambiente de teste separado (projeto Docker "analisefrota-e2e", banco próprio), roda testes/e2e.sh e apaga tudo.
# Não toca no banco de uso nem manda e-mail. Uso, na pasta servidor/:  bash testes/rodar-e2e.sh
set -euo pipefail
cd "$(dirname "$0")/.."

export PORTA_LOCAL=18080
dc() { docker compose -p analisefrota-e2e -f docker-compose.yml -f testes/docker-compose.e2e.yml "$@"; }
trap 'dc down -v >/dev/null 2>&1 || true' EXIT

dc up -d --build db app >/dev/null
for _ in $(seq 1 60); do curl -sf "http://127.0.0.1:$PORTA_LOCAL/saude" >/dev/null && break; sleep 1; done

link=$(dc run --rm app -criar-admin "Admin E2E" -email admin.e2e@exemplo.com 2>/dev/null | tail -1)

export BASE="http://127.0.0.1:$PORTA_LOCAL" CONVITE_ADMIN="${link##*#}"
# o teste lê o código de acesso do log e vence sessões direto no banco de teste
export LOGS_CMD="docker compose -p analisefrota-e2e -f docker-compose.yml -f testes/docker-compose.e2e.yml logs --no-log-prefix app"
export SQL_CMD="docker compose -p analisefrota-e2e -f docker-compose.yml -f testes/docker-compose.e2e.yml exec -T db psql -U frota -d frota -qtAc"
sh testes/e2e.sh
