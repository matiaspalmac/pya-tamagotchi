# Deploy VPS (DigitalOcean + Student Pack)

## Stack

VPS Ubuntu 24.04 + Docker Compose + Caddy (HTTPS automático) + dominio Namecheap (.me free 1 año Student Pack).

## Pre-requisitos

- GitHub Student Developer Pack activo (https://education.github.com/pack)
- DigitalOcean cuenta con $200 crédito reclamado
- Dominio (Namecheap .me free vía Student Pack)

## Costo

| Recurso | Tier | $/mes | Crédito gasta |
|---------|------|-------|---------------|
| Droplet Basic Regular 2GB | s-1vcpu-2gb | $12 | 16 meses con $200 |
| Droplet Basic Regular 1GB | s-1vcpu-1gb | $6 | 33 meses con $200 (apretado) |
| Droplet Premium Intel 2GB | s-1vcpu-2gb-intel | $14 | 14 meses |

**Recomendado:** $12/mes 2GB → ~16 meses gratis. Después migras o pagas $12.

## Paso 1: Reclamar DigitalOcean $200

1. https://education.github.com/pack
2. Buscar "DigitalOcean" → "Get access"
3. Sigue link → crea cuenta DigitalOcean → ingresa código pack
4. Verifica $200 en Billing

## Paso 2: Crear Droplet

Dashboard DO → Create → Droplets:

- **Region:** São Paulo (más cerca Chile) o NYC
- **Image:** Ubuntu 24.04 LTS x64
- **Size:** Basic Regular **$12/mes (2GB / 1 vCPU / 50GB SSD)**
- **Authentication:** SSH Key (recomendado) — sube tu pubkey
- **Hostname:** `tama-vps`
- **Backups:** opcional (+20% costo)

Click Create. ~30s. Anota IP pública.

## Paso 3: DNS (Namecheap o cualquiera)

Si reclamaste `.me` Namecheap free:
1. Dashboard Namecheap → Domain List → Manage → Advanced DNS
2. Agrega A records:
   ```
   tama          → <VPS_IP>     TTL Auto
   api.tama      → <VPS_IP>     TTL Auto
   ```
   (subdomains, root opcional)

Espera 1-5min propagación.

## Paso 4: Bootstrap VPS

SSH como root:

```bash
ssh root@<VPS_IP>
```

Corre bootstrap (instala Docker + clona repo + configura systemd):

```bash
curl -fsSL https://raw.githubusercontent.com/matiaspalmac/pya-tamagotchi/main/scripts/vps-bootstrap.sh | bash
```

## Paso 5: Configurar `.env`

```bash
nano /opt/tamagotchi/.env
```

Genera secrets:
```bash
echo "JWT_SECRET=$(openssl rand -hex 32)"
echo "POSTGRES_PASSWORD=$(openssl rand -base64 24)"
```

Setea dominios:
```
ACME_EMAIL=tu-email@gmail.com
WEB_DOMAIN=tama.tudominio.me
API_DOMAIN=api.tama.tudominio.me
PUBLIC_WEB_URL=https://tama.tudominio.me
PUBLIC_API_URL=https://api.tama.tudominio.me
```

## Paso 6: Deploy

```bash
bash /opt/tamagotchi/scripts/vps-deploy.sh
```

Caddy genera certificados Let's Encrypt automáticamente (~30s primera vez).

Verifica:
```bash
docker compose -f /opt/tamagotchi/deploy/prod/docker-compose.yml ps
curl https://api.tama.tudominio.me/healthz
```

Visita https://tama.tudominio.me en browser.

## Updates

Push a `main` → SSH a VPS → corre script:

```bash
ssh root@<VPS_IP> 'bash /opt/tamagotchi/scripts/vps-deploy.sh'
```

Auto-update cron opcional:
```bash
echo "0 4 * * * /opt/tamagotchi/scripts/vps-deploy.sh >> /var/log/tama-deploy.log 2>&1" | crontab -
```

## Logs

```bash
cd /opt/tamagotchi/deploy/prod
docker compose logs -f               # todos
docker compose logs -f pet           # uno
```

## Troubleshooting

**"Connection refused" en https://**
- Caddy aún emitiendo cert. Espera 1min, retry.
- Verifica DNS: `dig tama.tudominio.me`

**Out of memory**
- Droplet 1GB es justo. Sube a 2GB ($12/mes).
- O reduce: en docker-compose limita RAM por svc:
  ```yaml
  deploy:
    resources:
      limits:
        memory: 64M
  ```

**Postgres no arranca**
```bash
docker compose logs postgres
# Si volumen corrupto: docker compose down -v && bootstrap de nuevo
```

**Migraciones fallan (relation exists)**
- Inocuo, scripts SQL son idempotentes con `IF NOT EXISTS`. Ignora.

## Backup

Postgres dump diario:
```bash
echo '0 3 * * * docker compose -f /opt/tamagotchi/deploy/prod/docker-compose.yml exec -T postgres pg_dump -U tama tama | gzip > /var/backups/tama-$(date +\%F).sql.gz' | crontab -
```

Snapshot droplet completo: DO Dashboard → Droplet → Snapshots ($0.05/GB/mes).

## Hardening adicional

- Cambia SSH port + disable root login (`/etc/ssh/sshd_config`)
- Crea user no-root pa SSH
- Fail2ban ya instalado por bootstrap
- UFW solo abre 22, 80, 443

## Migración futura

Cuando $200 acabe, opciones:
- Hetzner CX22 €4.50/mes — mismo droplet más barato
- Pagar DO $12/mes
- Oracle Cloud Free permanente — migrar VM
