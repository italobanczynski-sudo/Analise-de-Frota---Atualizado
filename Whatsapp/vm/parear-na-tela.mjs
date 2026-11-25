// Pareia o WhatsApp da máquina virtual mostrando o QR Code numa página do navegador DESTE computador.
// Uso, na pasta robo-whatsapp:  node vm/parear-na-tela.mjs ubuntu@IP_DA_VM C:\caminho\chave
// A página se atualiza sozinha quando o WhatsApp troca o código (a cada ~20 s).
import { spawn, exec } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import qrcode from 'qrcode-terminal';

const [destino, chave] = process.argv.slice(2);
if (!destino) { console.log('Uso: node vm/parear-na-tela.mjs ubuntu@IP_DA_VM [chave]'); process.exit(1); }
const pagina = path.join(os.tmpdir(), 'robo-whatsapp-qr.html');
let aberta = false;

function escrever(corpo, atualizar = true) {
  fs.writeFileSync(pagina, `<!doctype html><meta charset="utf-8">${atualizar ? '<meta http-equiv="refresh" content="2">' : ''}
<title>Parear robô do WhatsApp</title>
<body style="margin:0;background:#fff;color:#111;font:16px system-ui;display:flex;flex-direction:column;align-items:center;padding:24px">
${corpo}</body>`);
  if (!aberta) { aberta = true; exec(`start "" "${pagina}"`); }
}
escrever('<h2>Conectando à máquina virtual…</h2><p>O QR Code aparece aqui em alguns segundos.</p>');

const args = [...(chave ? ['-i', chave] : []), '-o', 'BatchMode=yes', destino,
  "sudo -u robo -H env QR_BRUTO=1 bash -c 'cd /opt/robo-whatsapp && node robo.mjs parear'"];
const ssh = spawn('ssh', args, { stdio: ['ignore', 'pipe', 'inherit'] });
let resto = '';
ssh.stdout.on('data', (b) => {
  resto += b.toString();
  const linhas = resto.split('\n'); resto = linhas.pop();
  for (const l of linhas) {
    if (l.startsWith('QR:')) {
      qrcode.generate(l.slice(3).trim(), { small: false }, (txt) => {
        const bloco = txt.replace(/\x1b\[47m {2}\x1b\[0m/g, '<b style="background:#fff">  </b>').replace(/\x1b\[40m {2}\x1b\[0m/g, '<b style="background:#000">  </b>');
        escrever(`<h2>Leia com o celular do número do robô</h2>
<p>WhatsApp › Dispositivos conectados › Conectar um dispositivo</p>
<pre style="line-height:1;font-size:10px;font-family:monospace;margin:16px">${bloco}</pre>
<p style="color:#666">O código muda sozinho a cada ~20 s. Não feche esta página.</p>`);
      });
      console.log('QR Code novo na página do navegador');
    } else if (!/[▄█▀]/.test(l) && l.trim()) {
      console.log(l);
      if (/WhatsApp conectado/.test(l)) escrever('<h2 style="color:#0a7d32">✔ WhatsApp conectado!</h2><p>Pode fechar esta página.</p>', false);
      if (/ERRO/.test(l)) escrever(`<h2 style="color:#b00020">Não conectou</h2><p>${l.replace(/</g, '&lt;')}</p>`, false);
    }
  }
});
ssh.on('close', (c) => process.exit(c ?? 0));
