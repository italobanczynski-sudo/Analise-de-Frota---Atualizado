#!/usr/bin/env bash
# Troca o login do robô (ROBO_EMAIL e ROBO_SENHA) no .env da máquina virtual. A senha é digitada escondida.
# Uso:  sudo -u robo -H bash /opt/robo-whatsapp/vm/trocar-login.sh
set -euo pipefail
ARQ="$(cd "$(dirname "$0")/.." && pwd)/.env"
[ -f "$ARQ" ] || cp "$(dirname "$ARQ")/.env.example" "$ARQ"

read -rp 'E-mail do usuário do robô (@gobrax.com): ' EMAIL
[[ "$EMAIL" =~ @gobrax\.com$ ]] || { echo 'O e-mail precisa terminar em @gobrax.com.'; exit 1; }
read -rsp 'Senha do usuário do robô: ' SENHA; echo
[ -n "$SENHA" ] || { echo 'Senha vazia: nada mudou.'; exit 1; }

TMP="$(mktemp)"
grep -vE '^\s*ROBO_(EMAIL|SENHA)\s*=' "$ARQ" > "$TMP" || true
printf 'ROBO_EMAIL=%s\nROBO_SENHA=%s\n' "$EMAIL" "$SENHA" >> "$TMP"
install -m 600 "$TMP" "$ARQ" && rm -f "$TMP"
echo 'Login do robô atualizado. Teste com: npm run teste'
