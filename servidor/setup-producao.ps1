# ============================================================
# setup-producao.ps1
# Roda tudo para colocar a plataforma no ar com SSL e email
# ============================================================

$dominio = "analisefrota.gobrax.com.br"
$envFile = Join-Path $PSScriptRoot ".env"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Setup de Producao - Analise de Frota  " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# ── 1. Checar se o .env tem email preenchido ──────────────────
$envContent = Get-Content $envFile -Raw
if ($envContent -match "SEU_EMAIL@gmail.com" -or $envContent -match "SENHA_DE_APP_16_CHARS") {
    Write-Host "ATENCAO: O arquivo .env ainda tem placeholders nao preenchidos!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Abra o arquivo .env e substitua:" -ForegroundColor Yellow
    Write-Host "  SMTP_USUARIO=SEU_EMAIL@gmail.com   -> seu email real" -ForegroundColor Yellow
    Write-Host "  SMTP_SENHA=SENHA_DE_APP_16_CHARS   -> a senha de app gerada" -ForegroundColor Yellow
    Write-Host "  SMTP_REMETENTE=SEU_EMAIL@gmail.com -> seu email real" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Depois rode este script novamente." -ForegroundColor Yellow
    exit 1
}

Write-Host "[1/5] Subindo banco + app + nginx (HTTP)..." -ForegroundColor Green

# Usa configuracao HTTP-only temporariamente para o Certbot validar
# Remove a conf SSL se existir para nao travar o nginx sem certificado ainda
if (Test-Path "$PSScriptRoot\nginx\analisefrota.conf") {
    Rename-Item "$PSScriptRoot\nginx\analisefrota.conf" "$PSScriptRoot\nginx\analisefrota.conf.bak" -Force
}

docker compose up -d

Write-Host "[2/5] Aguardando nginx ficar pronto..." -ForegroundColor Green
Start-Sleep -Seconds 5

# ── 2. Gerar certificado SSL com Certbot ─────────────────────
Write-Host "[3/5] Gerando certificado SSL para $dominio..." -ForegroundColor Green
Write-Host "      (Certbot vai validar pelo dominio — precisa que o DNS aponte para este servidor)" -ForegroundColor DarkGray

$emailCertbot = (Select-String -Path $envFile -Pattern "SMTP_USUARIO=(.+)").Matches[0].Groups[1].Value.Trim()

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
    exit 1
}

# ── 3. Ativar config HTTPS no nginx ──────────────────────────
Write-Host "[4/5] Ativando HTTPS no nginx..." -ForegroundColor Green

if (Test-Path "$PSScriptRoot\nginx\analisefrota.conf.bak") {
    Rename-Item "$PSScriptRoot\nginx\analisefrota.conf.bak" "$PSScriptRoot\nginx\analisefrota.conf" -Force
}

# Restart nginx para carregar a config HTTPS
docker compose restart nginx

Write-Host "[5/5] Rebuild do app com configuracoes de producao..." -ForegroundColor Green
docker compose up -d --build app

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  PRONTO! Plataforma no ar!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "  URL: https://$dominio" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Verifique os logs com:" -ForegroundColor DarkGray
Write-Host "  docker compose logs -f app" -ForegroundColor DarkGray
Write-Host ""
