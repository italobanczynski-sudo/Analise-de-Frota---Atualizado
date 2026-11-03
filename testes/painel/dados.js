// Gera uma cópia do index.html com dados sintéticos (motoristas "Teste A".."Teste H", placas AAA1..AAA8; nada de cliente real)
// uso: node dados.js <index.html> <saida.html> <semNota 0|1> [comCliente 0|1]
// comCliente=1: cadastra o cliente "Teste Cliente" (local, com metas e tratativas), põe a coluna cliente e grava a planilha de veículos
const fs = require('fs');
const [src, out, semNota, comCliente] = process.argv.slice(2);
const cli = comCliente === '1';
const cols = ['data', 'motorista', 'veiculo', 'km', 'litros', 'frenagens', 'frenagens_alta', 'velocidade_media', 'odometro', 'horas', 'nota', 'horas_motor', 'horas_pedal',
  'faixa_verde', 'inicio_faixa_verde', 'final_faixa_verde', 'eco_roll', 'pressao_ideal', 'pressao_atencao', 'embalo', 'motor_ocioso', 'pressao_acelerador', 'excesso_rotacao', 'tempo_excesso_velocidade', 'piloto_automatico', 'freio_motor']
  .concat(cli ? ['cliente', 'grupo'] : []);
const linhas = [cols.join(';')];
let seed = 7; const r = () => ((seed = (seed * 16807) % 2147483647) / 2147483647);
const n = (v) => String(Math.round(v * 100) / 100).replace('.', ',');
const ini = new Date('2026-09-30T12:00:00Z');
for (let d = 0; d < 30; d++) {
  const dt = new Date(ini.getTime() + d * 864e5), br = dt.toISOString().slice(0, 10).split('-').reverse().join('/');
  for (let m = 1; m <= 8; m++) {
    // AAA8 para de rodar no 11º dia (caminhão parado ou unidade sem comunicar): registros com 0 km
    const parado = m === 8 && d >= 10;
    const km = parado ? 0 : 300 + r() * 300, kml = 2.3 + r() * 0.6 + d * 0.005;
    const v = [br, 'Teste ' + String.fromCharCode(64 + m), 'AAA' + m, n(km), n(km / kml), n(r() * 20), n(r() * 3), n(55 + r() * 10), n(100000 + d * 400), parado ? '0' : n(8 + r() * 2),
      semNota === '1' ? '' : n(50 + r() * 40), n(9 + r() * 2), n(6 + r() * 2)];
    for (let k = 13; k < 26; k++) v.push(n(r() * 60));
    if (cli) v.push('Teste Cliente', m <= 4 ? 'Grupo Norte' : 'Grupo Sul');
    linhas.push(v.join(';'));
  }
}
const csv = linhas.join('\n');
const sets = [['gbx-frota-csv', JSON.stringify({ nome: 'teste.csv', texto: csv })], ['gbx-reset-2026-10-21', '1']];
if (cli) {
  sets.push(['gbx-frota-veic', JSON.stringify({ nome: 'veiculos.csv', texto: csv })]);
  sets.push(['gbx-frota-clientes', JSON.stringify([{ id: 'l1', nome: 'Teste Cliente', grupo: '', obs: '', criado: '2026-09-01', metas: { nota: 75, kml: 2.6, motor_ocioso: 25 } }])]);
  sets.push(['gbx-frota-cliente-sel', 'Teste Cliente']);
  const et = ['estudar', 'plataforma', 'resultado'], dias = ['2026-10-02', '2026-10-09', '2026-10-15', '2026-10-15', '2026-10-16', '2026-10-21', '2026-10-21', '2026-10-22', '2026-10-27'];
  sets.push(['gbx-frota-tratativas', JSON.stringify(dias.map((d, i) => ({ id: 't' + i, cliente: 'Teste Cliente', data: d, etapa: et[i % 3], titulo: 'Teste ' + i, combinado: '', prazo: '', status: 'aberta', criado: d })))]);
}
let html = fs.readFileSync(src, 'utf8');
const pre = `<script>try{${sets.map(([k, v]) => `localStorage.setItem(${JSON.stringify(k)},${JSON.stringify(v)});`).join('')}}catch(e){}
window.addEventListener('error',function(e){document.documentElement.setAttribute('data-erros',(document.documentElement.getAttribute('data-erros')||'')+' | '+e.message);});</script>`;
const i = html.indexOf('<script');
html = html.slice(0, i) + pre + '\n' + html.slice(i);
fs.writeFileSync(out, html);
