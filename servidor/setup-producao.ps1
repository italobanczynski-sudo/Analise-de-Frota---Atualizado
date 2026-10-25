# ============================================================
# setup-producao.ps1
# Coloca a plataforma no ar com HTTPS (Let's Encrypt) e e-mail.
# Pré-requisitos: Docker, DNS do domínio apontando para esta máquina, portas 80 e 443 abertas, .env preenchido.
# ============================================================

$ErrorActionPreference = "Stop"
$dominio = "analisefrota.gobrax.com.br"
$envFile = Join-Path $PSScriptRoot ".env"
Set-Location $PSScriptRoot

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Setup de Producao - Analise de Frota  " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# ── 1. O .env existe e não tem mais os marcadores <...> do .env.example ──
if (-not (Test-Path $envFile)) {
    Write-Host "ERRO: copie .env.example para .env e preencha antes de rodar este script." -ForegroundColor Red
    exit 1
}
$pendentes = Select-String -Path $envFile -Pattern "^[A-Z_]+=.*<[^>]+>" | ForEach-Object { ($_.Line -split "=")[0] }
if ($pendentes) {
    Write-Host "ATENCAO: o .env ainda tem campos com <marcador>: $($pendentes -join ', ')" -ForegroundColor Red
    Write-Host "Preencha esses campos e rode o script de novo." -ForegroundColor Yellow
    exit 1
}
if (-not (Select-String -Path $envFile -Pattern "^URL_PUBLICA=https://$dominio" -Quiet)) {
    Write-Host "ATENCAO: URL_PUBLICA no .env precisa ser https://$dominio" -ForegroundColor Red
    exit 1
}

# ── 2. Sobe banco + app + nginx só em HTTP (o Certbot valida o domínio pela porta 80) ──
Write-Host "[1/4] Subindo banco, app e nginx em HTTP..." -ForegroundColor Green
$env:NGINX_SITE = "http-somente"
docker compose --profile producao up -d --build
if ($LASTEXITCODE -ne 0) { exit 1 }
Start-Sleep -Seconds 5

# ── 3. Certificado SSL ──
Write-Host "[2/4] Gerando certificado SSL para $dominio..." -ForegroundColor Green
$emailCertbot = (Select-String -Path $envFile -Pattern "^SMTP_USUARIO=(.+)").Matches[0].Groups[1].Value.Trim()
docker compose --profile certbot run --rm certbot certonly `
    --webroot -w /var/www/certbot `
    -d $dominio `
    --email $emailCertbot `
    --agree-tos --no-eff-email --non-interactive

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "ERRO: Certbot falhou. Verifique se:" -ForegroundColor Red
    Write-Host "  - O DNS de $dominio aponta para o IP deste servidor" -ForegroundColor Yellow
    Write-Host "  - As portas 80 e 443 estao abertas no firewall" -ForegroundColor Yellow
    Write-Host "Enquanto isso o nginx fica em HTTP e o LOGIN NAO FUNCIONA (o cookie de sessao exige HTTPS)." -ForegroundColor Yellow
    exit 1
}

# ── 4. Troca o nginx para o site com HTTPS (o padrão do docker-compose) ──
Write-Host "[3/4] Ativando HTTPS no nginx..." -ForegroundColor Green
Remove-Item Env:NGINX_SITE
docker compose --profile producao up -d --force-recreate nginx
if ($LASTEXITCODE -ne 0) { exit 1 }

Write-Host "[4/4] Conferindo..." -ForegroundColor Green
Start-Sleep -Seconds 3
try {
    $r = Invoke-WebRequest -UseBasicParsing "https://$dominio/saude" -TimeoutSec 15
    Write-Host "  /saude respondeu $($r.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "  /saude nao respondeu por HTTPS: veja 'docker compose --profile producao logs nginx'" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  PRONTO! Plataforma no ar!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "  URL: https://$dominio" -ForegroundColor Cyan
Write-Host "  Primeiro administrador (se ainda nao existir):" -ForegroundColor DarkGray
Write-Host "  docker compose run --rm app -criar-admin `"Nome`" -email pessoa@gobrax.com" -ForegroundColor DarkGray
Write-Host "  Logs: docker compose logs -f app" -ForegroundColor DarkGray
Write-Host ""
