#!/usr/bin/env bash
# Copia o banco deste computador (docker compose) para um Postgres na nuvem (Render, Neon...).
# A URL do banco de destino fica só no arquivo .banco-nuvem (fora do Git): cole nele a "External Database URL".
# Apaga e recria as tabelas no destino. Uso, na pasta servidor/:  bash copiar-dados-nuvem.sh
set -euo pipefail
cd "$(dirname "$0")"
URL=$(tr -d '\r\n' < .banco-nuvem)
case "$URL" in postgres://*|postgresql://*) ;; *) echo "Cole a External Database URL no arquivo servidor/.banco-nuvem"; exit 1;; esac
case "$URL" in *sslmode=*) ;; *\?*) URL="$URL&sslmode=require";; *) URL="$URL?sslmode=require";; esac
docker compose exec -T db pg_dump -U frota -d frota --clean --if-exists --no-owner --no-privileges \
  | docker run --rm -i -e URL="$URL" postgres:16-alpine sh -c 'psql "$URL" -q -v ON_ERROR_STOP=0 >/dev/null'
docker run --rm -e URL="$URL" postgres:16-alpine psql "$URL" -tAc \
  "SELECT 'usuarios '||(SELECT count(*) FROM usuarios)||', clientes '||(SELECT count(*) FROM clientes)||', registros '||(SELECT count(*) FROM registros_diarios)||', documentos do painel '||(SELECT count(*) FROM base_documentos)"
