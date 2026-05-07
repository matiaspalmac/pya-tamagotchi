#!/usr/bin/env bash
# Bootstrap VPS Ubuntu 24.04 — corre como root o con sudo.
# Uso: curl -fsSL https://raw.githubusercontent.com/matiaspalmac/pya-tamagotchi/main/scripts/vps-bootstrap.sh | bash
set -euo pipefail

REPO_URL="${REPO_URL:-https://github.com/matiaspalmac/pya-tamagotchi.git}"
APP_DIR="${APP_DIR:-/opt/tamagotchi}"
APP_USER="${APP_USER:-tama}"

echo "==> 1. Update base"
apt-get update -qq
apt-get upgrade -y -qq

echo "==> 2. Paquetes necesarios"
apt-get install -y -qq ca-certificates curl gnupg git ufw fail2ban

echo "==> 3. Docker oficial"
if ! command -v docker >/dev/null; then
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  chmod a+r /etc/apt/keyrings/docker.asc
  source /etc/os-release
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $VERSION_CODENAME stable" \
    | tee /etc/apt/sources.list.d/docker.list >/dev/null
  apt-get update -qq
  apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
fi

echo "==> 4. Usuario app"
if ! id "$APP_USER" >/dev/null 2>&1; then
  useradd -m -s /bin/bash "$APP_USER"
  usermod -aG docker "$APP_USER"
fi

echo "==> 5. Firewall"
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 443/udp
ufw --force enable

echo "==> 6. Clone repo en $APP_DIR"
if [ ! -d "$APP_DIR/.git" ]; then
  mkdir -p "$APP_DIR"
  git clone "$REPO_URL" "$APP_DIR"
  chown -R "$APP_USER:$APP_USER" "$APP_DIR"
fi

echo "==> 7. .env (si no existe)"
if [ ! -f "$APP_DIR/.env" ]; then
  cp "$APP_DIR/deploy/prod/.env.example" "$APP_DIR/.env"
  chown "$APP_USER:$APP_USER" "$APP_DIR/.env"
  chmod 600 "$APP_DIR/.env"
  echo "==> ⚠️  EDITA $APP_DIR/.env antes de continuar"
  echo "    Genera secrets:"
  echo "      JWT_SECRET=\$(openssl rand -hex 32)"
  echo "      POSTGRES_PASSWORD=\$(openssl rand -base64 24)"
fi

echo "==> 8. Systemd unit pa auto-start"
cat >/etc/systemd/system/tamagotchi.service <<EOF
[Unit]
Description=Tamagotchi Multiplayer Stack
Requires=docker.service
After=docker.service network-online.target
Wants=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$APP_DIR/deploy/prod
EnvironmentFile=$APP_DIR/.env
ExecStart=/usr/bin/docker compose up -d
ExecStop=/usr/bin/docker compose down
TimeoutStartSec=600

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable tamagotchi.service

echo
echo "==> Bootstrap completo."
echo "   1. Edita $APP_DIR/.env con tus secrets + dominios"
echo "   2. Apunta DNS A: <VPS_IP> → tudominio + api.tudominio"
echo "   3. Corre: bash $APP_DIR/scripts/vps-deploy.sh"
