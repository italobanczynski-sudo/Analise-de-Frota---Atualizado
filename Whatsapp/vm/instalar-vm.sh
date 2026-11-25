#!/usr/bin/env bash
# Instala o robô do WhatsApp numa máquina virtual Ubuntu (22.04 ou 24.04) e agenda para todo dia às 08:00 (Brasília).
# Roda de novo sem problema: serve também para atualizar o código (o .env e a sessão do WhatsApp ficam).
#
# Uso, na máquina virtual, dentro da pasta robo-whatsapp que veio do enviar-para-vm.ps1:
#   sudo bash vm/instalar-vm.sh            (horário padrão 08:00)
#   sudo HORA=07:30 bash vm/instalar-vm.sh
set -euo pipefail

HORA="${HORA:-08:00}"
USUARIO=robo
DESTINO=/opt/robo-whatsapp
ORIGEM="$(cd "$(dirname "$0")/.." && pwd)"

[ "$(id -u)" = 0 ] || { echo "Rode com sudo: sudo bash vm/instalar-vm.sh"; exit 1; }
export DEBIAN_FRONTEND=noninteractive

echo "[1/6] Fuso horário de Brasília"
timedatectl set-timezone America/Sao_Paulo

echo "[2/6] Memória: com pouca RAM, cria 2 GB de swap"
if [ "$(awk '/MemTotal/ {print $2}' /proc/meminfo)" -lt 1900000 ] && ! swapon --show | grep -q /swapfile; then
  fallocate -l 2G /swapfile && chmod 600 /swapfile && mkswap /swapfile >/dev/null && swapon /swapfile
  grep -q '^/swapfile' /etc/fstab || echo '/swapfile none swap sw 0 0' >> /etc/fstab
fi

echo "[3/6] Node.js 20"
apt-get update -qq
apt-get install -y -qq ca-certificates curl gnupg rsync fonts-noto-color-emoji >/dev/null
if ! command -v node >/dev/null || [ "$(node -p 'process.versions.node.split(".")[0]')" -lt 20 ]; then
  curl -fsSL https://deb.nodesource.com/setup_20.x | bash - >/dev/null
  apt-get install -y -qq nodejs >/dev/null
fi

echo "[4/6] Usuário '$USUARIO' e código em $DESTINO"
id "$USUARIO" >/dev/null 2>&1 || useradd --create-home --shell /bin/bash "$USUARIO"
mkdir -p "$DESTINO"
# .env, sessão e logs da máquina virtual nunca são sobrescritos
rsync -a --delete --exclude node_modules --exclude .env --exclude logs --exclude .sessao-whatsapp --exclude .wwebjs_cache \
  "$ORIGEM"/ "$DESTINO"/
if [ ! -f "$DESTINO/.env" ]; then
  grep -v '^CHROME=' "$DESTINO/.env.example" > "$DESTINO/.env"
  chmod 600 "$DESTINO/.env"
fi
chown -R "$USUARIO:$USUARIO" "$DESTINO"
sudo -u "$USUARIO" -H bash -c "cd '$DESTINO' && npm ci --omit=dev --no-audit --no-fund --loglevel=error"

echo "[5/6] Agendamento (systemd): todo dia às $HORA"
cat > /etc/systemd/system/robo-whatsapp.service <<EOF
[Unit]
Description=Robô do WhatsApp Gobrax: manda as Mensagens do dia aos motoristas
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
User=$USUARIO
WorkingDirectory=$DESTINO
ExecStart=/usr/bin/node $DESTINO/robo.mjs
TimeoutStartSec=5h
EOF
cat > /etc/systemd/system/robo-whatsapp.timer <<EOF
[Unit]
Description=Roda o robô do WhatsApp todo dia às $HORA (Brasília)

[Timer]
OnCalendar=*-*-* $HORA:00 America/Sao_Paulo
# se a máquina estava desligada no horário, roda assim que ligar
Persistent=true

[Install]
WantedBy=timers.target
EOF
systemctl daemon-reload
systemctl enable --now robo-whatsapp.timer >/dev/null

echo "[6/6] Pronto. Próximo envio:"
systemctl list-timers robo-whatsapp.timer --no-pager | sed -n 2p

cat <<EOF

Falta (só na primeira vez):
  1. Login do robô na plataforma:  sudo -u $USUARIO -H bash $DESTINO/vm/trocar-login.sh
  2. Parear o WhatsApp (QR Code aparece aqui no terminal):
       sudo -u $USUARIO -H bash -c 'cd $DESTINO && npm run parear'
  3. Conferir sem enviar:          sudo -u $USUARIO -H bash -c 'cd $DESTINO && npm run teste'

Dia a dia:  logs em $DESTINO/logs  ·  enviar agora: sudo systemctl start robo-whatsapp
EOF
