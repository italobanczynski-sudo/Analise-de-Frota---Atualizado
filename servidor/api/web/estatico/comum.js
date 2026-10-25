// Funções comuns às telas: chamar a API do próprio servidor e avisar o resultado.
'use strict';

const PAPEIS = { admin: 'Administrador', colaborador: 'Colaborador', leitor: 'Leitor' };

// api chama o servidor com o cookie de login e o cabeçalho que a plataforma exige; devolve os dados ou lança o erro com a mensagem do servidor
async function api(metodo, caminho, corpo) {
  const opcoes = { method: metodo, credentials: 'same-origin', headers: { 'X-Requisicao': 'plataforma' } };
  if (corpo !== undefined) {
    opcoes.headers['Content-Type'] = 'application/json';
    opcoes.body = JSON.stringify(corpo);
  }
  let r;
  try {
    r = await fetch(caminho, opcoes);
  } catch (e) {
    throw new Error('Sem conexão com o servidor. Confira se ele está ligado e tente de novo.');
  }
  let j = null;
  try { j = await r.json(); } catch (e) { /* resposta sem JSON */ }
  if (!r.ok || !j || !j.ok) {
    const erro = new Error((j && j.erro) || 'O servidor respondeu ' + r.status + '.');
    erro.status = r.status;
    throw erro;
  }
  return j.dados;
}

// irParaPlataforma vai para a área logada só depois de conferir que o navegador guardou o login.
// Com URL_PUBLICA em https, o cookie é "Secure": numa página aberta por http:// o navegador o descarta e a
// pessoa voltaria para a tela de entrada sem nenhum aviso.
async function irParaPlataforma() {
  try {
    await api('GET', '/api/v1/auth/eu');
  } catch (e) {
    throw new Error(location.protocol === 'http:'
      ? 'A senha foi aceita, mas o navegador não guardou o login porque esta página foi aberta sem conexão segura (http://). Abra o endereço com https:// e entre de novo.'
      : 'A senha foi aceita, mas o navegador não guardou o login. Libere os cookies deste site e entre de novo.');
  }
  location.replace('/plataforma');
}

function toast(texto, erro) {
  document.querySelectorAll('.toast').forEach((t) => t.remove());
  const t = document.createElement('div');
  t.className = 'toast' + (erro ? ' erro' : '');
  t.setAttribute('role', 'status');
  t.textContent = texto;
  document.body.appendChild(t);
  setTimeout(() => t.remove(), erro ? 7000 : 3500);
}

function mostrarErro(el, texto) {
  el.textContent = texto;
  el.hidden = !texto;
}

function iniciais(nome) {
  return String(nome || '?').trim().split(/\s+/).filter(Boolean).slice(0, 2).map((p) => p[0].toUpperCase()).join('') || '?';
}

function dataHora(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  return d.toLocaleDateString('pt-BR') + ' ' + d.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });
}

function diaBR(iso) {
  if (!iso) return '—';
  const [a, m, d] = String(iso).slice(0, 10).split('-');
  return d + '/' + m + '/' + a;
}

// el cria um elemento com classe, texto e filhos (sem innerHTML: nada que venha da API vira código na página)
function el(tag, props, ...filhos) {
  const e = document.createElement(tag);
  Object.entries(props || {}).forEach(([k, v]) => {
    if (k === 'class') e.className = v;
    else if (k === 'text') e.textContent = v;
    else if (k.startsWith('on')) e.addEventListener(k.slice(2), v);
    else if (v !== false && v !== null && v !== undefined) e.setAttribute(k, v === true ? '' : v);
  });
  filhos.flat().forEach((f) => { if (f !== null && f !== undefined && f !== false) e.append(f); });
  return e;
}
