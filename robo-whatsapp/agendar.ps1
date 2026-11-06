# Agenda o robô no Agendador de Tarefas do Windows: todo dia às 08:00 (ou assim que o computador ligar, se estava desligado).
# Uso, no PowerShell desta pasta:  powershell -ExecutionPolicy Bypass -File agendar.ps1
# Para tirar o agendamento:         Unregister-ScheduledTask -TaskName "Robo WhatsApp Gobrax" -Confirm:$false
param([string]$Hora = '08:00')

$pasta = $PSScriptRoot
$node = (Get-Command node -ErrorAction Stop).Source
$acao = New-ScheduledTaskAction -Execute $node -Argument "`"$pasta\robo.mjs`"" -WorkingDirectory $pasta
$quando = New-ScheduledTaskTrigger -Daily -At $Hora
$ajustes = New-ScheduledTaskSettingsSet -StartWhenAvailable -WakeToRun -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit (New-TimeSpan -Hours 5) -MultipleInstances IgnoreNew
Register-ScheduledTask -TaskName 'Robo WhatsApp Gobrax' -Action $acao -Trigger $quando -Settings $ajustes `
  -Description 'Manda as Mensagens do dia aos motoristas pelo WhatsApp Web (robo-whatsapp/robo.mjs)' -Force | Out-Null
Write-Host "Agendado: todo dia as $Hora. Log em $pasta\logs."
