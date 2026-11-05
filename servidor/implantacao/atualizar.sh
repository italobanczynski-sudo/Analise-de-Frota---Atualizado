#!/usr/bin/env bash
# Atualiza a plataforma quando chega código novo no GitHub (branch main). Sem novidade, não faz nada.
# Agende no servidor para rodar a cada 5 minutos (crontab -e):
#   */5 * * * * /caminho/do/repositorio/servidor/implantacao/atualizar.sh >> /var/log/analisefrota-atualizar.log 2>&1
set -euo pipefail
cd "$(dirname "$0")/../.."
git fetch -q origin main
if [ "$(git rev-parse HEAD)" = "$(git rev-parse origin/main)" ]; then exit 0; fi
echo "$(date '+%F %T') atualizando para $(git rev-parse --short origin/main)"
git merge -q --ff-only origin/main
cd servidor
docker compose --profile producao up -d --build
docker image prune -f >/dev/null
