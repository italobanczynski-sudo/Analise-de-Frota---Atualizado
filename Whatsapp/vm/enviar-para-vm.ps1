# Manda o código do robô deste computador para a máquina virtual e instala/atualiza lá (vm/instalar-vm.sh).
# Não manda o .env, a sessão do WhatsApp nem os logs: o login é feito na própria máquina virtual (vm/trocar-login.sh).
# Uso, no PowerShell da pasta robo-whatsapp:
#   powershell -ExecutionPolicy Bypass -File vm\enviar-para-vm.ps1 -Destino ubuntu@IP_DA_VM [-Chave C:\caminho\chave.pem]
param(
  [Parameter(Mandatory = $true)][string]$Destino,
  [string]$Chave = ''
)
$ErrorActionPreference = 'Stop'
$pasta = Split-Path $PSScriptRoot -Parent
$pacote = Join-Path $env:TEMP 'robo-whatsapp.tgz'
$opcoes = @()
if ($Chave) { $opcoes = @('-i', $Chave) }

Write-Host '[1/3] Empacotando o código (sem .env, sessão, logs e node_modules)...'
tar -czf $pacote -C $pasta --exclude=node_modules --exclude=.env --exclude=logs --exclude=.sessao-whatsapp --exclude=.wwebjs_cache .
if ($LASTEXITCODE -ne 0) { exit 1 }

Write-Host "[2/3] Enviando para $Destino..."
scp @opcoes $pacote "${Destino}:/tmp/robo-whatsapp.tgz"
if ($LASTEXITCODE -ne 0) { exit 1 }

Write-Host '[3/3] Instalando na máquina virtual (pede a senha do sudo, se houver)...'
ssh -t @opcoes $Destino 'rm -rf /tmp/robo-whatsapp && mkdir -p /tmp/robo-whatsapp && tar -xzf /tmp/robo-whatsapp.tgz -C /tmp/robo-whatsapp && sudo bash /tmp/robo-whatsapp/vm/instalar-vm.sh'
Remove-Item $pacote -ErrorAction SilentlyContinue
