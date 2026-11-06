# Robô do WhatsApp

Todo dia às 08:00 manda a análise de ontem (Mensagens do dia) a cada motorista **com telefone cadastrado** e marca como enviada na plataforma. Roda neste computador, com o WhatsApp Web aberto por baixo, sem janela.

> **Risco:** automatizar o WhatsApp Web é contra os Termos do WhatsApp e o número pode ser bloqueado. Use um número só para o robô, nunca o pessoal nem o da empresa. As pausas aleatórias entre mensagens (25–60 s) diminuem o risco, não o eliminam.

## Instalar (uma vez)

1. Na plataforma, em **Configurações › Usuários**, crie um usuário para o robô com papel **Colaborador** e crie a senha pelo convite.
2. Nesta pasta: `npm install`.
3. Copie `.env.example` para `.env` e preencha `ROBO_EMAIL` e `ROBO_SENHA`. O `.env` não vai para o Git.
4. `npm run parear`: abre o Chrome com o QR Code. No celular do número do robô: **WhatsApp › Dispositivos conectados › Conectar um dispositivo**. A sessão fica em `.sessao-whatsapp/`.
5. `npm run teste`: mostra quem receberia a mensagem hoje, sem enviar nada.
6. `powershell -ExecutionPolicy Bypass -File agendar.ps1`: agenda no Windows para todo dia às 08:00. Se o computador estiver desligado nesse horário, roda quando ligar.

## No dia a dia

- Só recebe quem tem telefone em **Mensagens do dia** (campo em cada motorista). Quem já foi marcado como enviado não recebe de novo.
- Se às 08:00 os dados de ontem ainda não chegaram, o robô confere de novo a cada 20 min até 12:00.
- O que aconteceu fica em `logs/AAAA-MM-DD.log` (telefones mascarados).
- O computador precisa estar ligado e com internet. Se o WhatsApp desconectar o aparelho (ex.: celular sem abrir o WhatsApp por 14 dias), rode `npm run parear` de novo.
- Para enviar agora, sem esperar o horário: `npm run enviar`. Para tirar o agendamento: `Unregister-ScheduledTask -TaskName "Robo WhatsApp Gobrax" -Confirm:$false`.
