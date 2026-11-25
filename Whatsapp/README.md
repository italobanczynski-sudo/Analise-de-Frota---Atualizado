# Robô do WhatsApp

Todo dia às 08:00 manda a análise de ontem (Mensagens do dia) a cada motorista **com telefone cadastrado** e marca como enviada na plataforma. Conecta ao WhatsApp como um aparelho a mais do número do robô (biblioteca Baileys, sem navegador). Roda numa máquina virtual na AWS (veja abaixo) ou neste computador.

> **Risco:** automatizar o WhatsApp é contra os Termos do WhatsApp e o número pode ser bloqueado. Use um número só para o robô, nunca o pessoal nem o da empresa. As pausas aleatórias entre mensagens (25–60 s) diminuem o risco, não o eliminam.

## Instalar (uma vez)

1. Na plataforma, em **Configurações › Usuários**, crie um usuário para o robô com papel **Colaborador** e crie a senha pelo convite.
2. Nesta pasta: `npm install`.
3. Copie `.env.example` para `.env` e preencha `ROBO_EMAIL` e `ROBO_SENHA`. O `.env` não vai para o Git.
4. `npm run parear`: mostra o QR Code no terminal. No celular do número do robô: **WhatsApp › Dispositivos conectados › Conectar um dispositivo**. A sessão fica em `%LOCALAPPDATA%\robo-whatsapp-gobrax\sessao` (no Linux, `~/.local/share/robo-whatsapp-gobrax/sessao`). Quem tiver essa pasta manda mensagem pelo número do robô: não copie nem compartilhe.
5. `npm run teste`: mostra quem receberia a mensagem hoje, sem enviar nada.
6. `powershell -ExecutionPolicy Bypass -File agendar.ps1`: agenda no Windows para todo dia às 08:00. Se o computador estiver desligado nesse horário, roda quando ligar.

## Rodar numa máquina virtual (sem depender deste computador ligado)

O robô também roda numa máquina virtual Linux na nuvem, ligada 24 h. Lá ele é agendado pelo `systemd` (todo dia às 08:00 de Brasília; se a máquina estava desligada no horário, roda quando ligar).

1. **Crie a máquina virtual** no provedor (AWS Lightsail, Google Cloud, Oracle Cloud, Hetzner, Azure…): **Ubuntu 22.04 ou 24.04, x86-64 (amd64), 2 GB de RAM** (com 1 GB o instalador cria swap, mas fica lento). Não precisa abrir porta nenhuma além do SSH (22). Guarde o IP e a chave SSH.
2. **Neste computador**, na pasta `robo-whatsapp`:
   `powershell -ExecutionPolicy Bypass -File vm\enviar-para-vm.ps1 -Destino ubuntu@IP_DA_VM -Chave C:\caminho\chave.pem`
   Ele manda o código (sem `.env`, sessão e logs) e roda `vm/instalar-vm.sh`, que instala o Node 20, põe o código em `/opt/robo-whatsapp`, cria o usuário `robo` e o agendamento.
3. **Na máquina virtual** (`ssh -i chave.pem ubuntu@IP_DA_VM`):
   - login do robô: `sudo -u robo -H bash /opt/robo-whatsapp/vm/trocar-login.sh`
   - parear: `sudo -u robo -H bash -c 'cd /opt/robo-whatsapp && npm run parear'`. O QR Code aparece no próprio terminal; leia com o celular do número do robô.
   - conferir: `sudo -u robo -H bash -c 'cd /opt/robo-whatsapp && npm run teste'`
4. **Desligue o agendamento deste computador**, senão os dois mandam: `Unregister-ScheduledTask -TaskName "Robo WhatsApp Gobrax" -Confirm:$false`. Desconecte também este computador no celular (WhatsApp › Dispositivos conectados).

Na máquina virtual: próximo horário `systemctl list-timers robo-whatsapp.timer`, enviar agora `sudo systemctl start robo-whatsapp`, acompanhar `journalctl -u robo-whatsapp -f`, logs em `/opt/robo-whatsapp/logs`. Para atualizar o código, rode o passo 2 de novo (o `.env` e a sessão ficam). Outro horário: `sudo HORA=07:30 bash /opt/robo-whatsapp/vm/instalar-vm.sh`.

## No dia a dia

- Só recebe quem tem telefone em **Mensagens do dia** (campo em cada motorista). Quem já foi marcado como enviado não recebe de novo.
- Se às 08:00 os dados de ontem ainda não chegaram, o robô confere de novo a cada 20 min até 12:00.
- O que aconteceu fica em `logs/AAAA-MM-DD.log` (telefones mascarados).
- A máquina que roda o robô (hoje, a da AWS) precisa estar ligada e com internet. Se o WhatsApp desconectar o aparelho (ex.: celular sem abrir o WhatsApp por 14 dias), rode `npm run parear` de novo.
- Para enviar agora, sem esperar o horário: `npm run enviar`. Para tirar o agendamento: `Unregister-ScheduledTask -TaskName "Robo WhatsApp Gobrax" -Confirm:$false`.
