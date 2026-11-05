#!/usr/bin/env bash
# Traz o banco do Render (ou de outro Postgres na nuvem) para o banco deste servidor (docker compose).
# Usuários, clientes, tokens cifrados, histórico, documentos do painel e rastro vêm juntos.
# Os tokens só abrem se a CHAVE_TOKENS do .env deste servidor for a mesma usada no Render.
# A URL do banco de origem fica só no arquivo servidor/.banco-render (fora do Git): cole nele a
# "External Database URL" do painel do Render (banco analise-de-frota-db → Connect).
# APAGA e recria as tabelas deste servidor. Uso, na pasta servidor/:  bash implantacao/trazer-dados-render.sh
set -euo pipefail
cd "$(dirname "$0")/.."
[ -f .banco-render ] || { echo "Crie o arquivo servidor/.banco-render com a External Database URL do Render"; exit 1; }
URL=$(tr -d '\r\n' < .banco-render)
case "$URL" in postgres://*|postgresql://*) ;; *) echo "O arquivo servidor/.banco-render não tem uma URL postgres://"; exit 1;; esac
case "$URL" in *sslmode=*) ;; *\?*) URL="$URL&sslmode=require";; *) URL="$URL?sslmode=require";; esac
read -r -p "Isto APAGA os dados deste servidor e coloca os do Render no lugar. Continuar? (digite sim) " ok
[ "$ok" = "sim" ] || { echo "Cancelado."; exit 1; }
echo "Guardando uma cópia do banco atual em backup-antes-render.dump..."
docker compose exec -T db pg_dump -U frota -d frota -Fc > backup-antes-render.dump
echo "Copiando do Render..."
docker run --rm -e URL="$URL" postgres:16-alpine pg_dump "$URL" --clean --if-exists --no-owner --no-privileges \
  | docker compose exec -T db psql -U frota -d frota -q -v ON_ERROR_STOP=0 >/dev/null
docker compose exec -T db psql -U frota -d frota -tAc \
  "SELECT 'usuarios '||(SELECT count(*) FROM usuarios)||', clientes '||(SELECT count(*) FROM clientes)||', registros '||(SELECT count(*) FROM registros_diarios)||', documentos do painel '||(SELECT count(*) FROM base_documentos)"
echo "Reiniciando o app..."
docker compose restart app
echo "Pronto. Apague o arquivo servidor/.banco-render (ele dá acesso ao banco do Render)."
