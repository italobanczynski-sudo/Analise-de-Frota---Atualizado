# Troca o login do robô (ROBO_EMAIL e ROBO_SENHA) no .env desta pasta. A senha é digitada escondida e não aparece na tela.
# Uso, no PowerShell desta pasta:  powershell -ExecutionPolicy Bypass -File trocar-login.ps1
$arq = Join-Path $PSScriptRoot '.env'
if (-not (Test-Path $arq)) { Copy-Item (Join-Path $PSScriptRoot '.env.example') $arq }

$email = (Read-Host 'E-mail do usuario do robo (@gobrax.com)').Trim()
if ($email -notmatch '@gobrax\.com$') { Write-Host 'O e-mail precisa terminar em @gobrax.com.'; exit 1 }
$seg = Read-Host 'Senha do usuario do robo' -AsSecureString
$senha = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($seg))
if (-not $senha) { Write-Host 'Senha vazia: nada mudou.'; exit 1 }

$linhas = Get-Content $arq -Encoding UTF8 | Where-Object { $_ -notmatch '^\s*ROBO_(EMAIL|SENHA)\s*=' }
$linhas += "ROBO_EMAIL=$email"
$linhas += "ROBO_SENHA=$senha"
[IO.File]::WriteAllLines($arq, $linhas, (New-Object Text.UTF8Encoding $false))
Write-Host 'Login do robo atualizado. Teste com: npm run teste'
