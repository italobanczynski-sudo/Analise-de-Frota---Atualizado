// Tela de entrada: login, primeiro acesso e senha esquecida (código por e-mail).
'use strict';

const $ = (id) => document.getElementById(id);
const emailValido = (e) => /^\S+@\S+\.\S+$/.test(e);

// já logado: vai direto para a plataforma
api('GET', '/api/v1/auth/eu').then(() => location.replace('/plataforma')).catch(() => {});

function aviso(id, texto, ok) {
  const m = $(id);
  m.textContent = texto || '';
  m.classList.toggle('ok', !!ok);
}

function carregando(botao, sim) {
  $(botao).classList.toggle('load', sim);
  $(botao).disabled = sim;
}

function erroCampo(id, sim) { $(id).classList.toggle('err', sim); }

// ---------- etapas do cartão ----------
let modo = 'primeiro';
function irPara(etapa) {
  ['entrar', 'pedir', 'codigo'].forEach((e) => { $('etapa-' + e).hidden = e !== etapa; });
  if (etapa === 'pedir') {
    $('titulo-pedir').textContent = modo === 'esqueci' ? 'Recuperar senha' : 'Primeiro acesso';
    $('sub-pedir').textContent = modo === 'esqueci'
      ? 'Digite seu e-mail. Enviaremos um código de 6 dígitos para você criar uma senha nova.'
      : 'Digite o e-mail que a Gobrax cadastrou para você. Enviaremos um código de 6 dígitos para criar sua senha.';
    if (!$('email2').value) $('email2').value = $('email').value.trim();
    $('email2').focus();
  }
  if (etapa === 'entrar') $('email').focus();
  if (etapa === 'codigo') $('codigo').focus();
}
document.querySelectorAll('[data-ir]').forEach((b) => b.addEventListener('click', () => {
  if (b.dataset.ir !== 'entrar') modo = b.dataset.ir;
  irPara(b.dataset.ir === 'entrar' ? 'entrar' : 'pedir');
}));

// mostrar/ocultar senha
document.querySelectorAll('.eye').forEach((b) => b.addEventListener('click', () => {
  const p = $(b.dataset.alvo), oculto = p.type === 'password';
  p.type = oculto ? 'text' : 'password';
  b.setAttribute('aria-label', oculto ? 'Ocultar senha' : 'Mostrar senha');
}));

// ---------- 1. entrar ----------
$('form-entrar').addEventListener('submit', async (ev) => {
  ev.preventDefault();
  const email = $('email').value.trim(), senha = $('pass').value;
  erroCampo('fe', !emailValido(email));
  erroCampo('fp', !senha);
  if (!emailValido(email)) return aviso('msg-entrar', 'Informe um e-mail válido.');
  if (!senha) return aviso('msg-entrar', 'Informe sua senha.');
  aviso('msg-entrar', '');
  carregando('go', true);
  try {
    await api('POST', '/api/v1/auth/entrar', { email, senha });
    await irParaPlataforma();
  } catch (e) {
    carregando('go', false);
    erroCampo('fp', true);
    aviso('msg-entrar', e.message);
    $('pass').select();
  }
});

// ---------- 2. pedir o código ----------
let emailCodigo = '', espera = null;
async function pedirCodigo(email, botao, msgId) {
  carregando(botao, true);
  try {
    await api('POST', '/api/v1/auth/codigo', { email });
    emailCodigo = email;
    $('email-enviado').textContent = email;
    irPara('codigo');
    aviso('msg-codigo', 'Código enviado. Ele vale por 15 minutos.', true);
    contarReenvio();
  } catch (e) {
    aviso(msgId, e.message);
  }
  carregando(botao, false);
}

$('form-pedir').addEventListener('submit', (ev) => {
  ev.preventDefault();
  const email = $('email2').value.trim();
  erroCampo('fe2', !emailValido(email));
  if (!emailValido(email)) return aviso('msg-pedir', 'Informe um e-mail válido.');
  aviso('msg-pedir', '');
  pedirCodigo(email, 'go-pedir', 'msg-pedir');
});

// um código novo só depois de 60 s
function contarReenvio() {
  let s = 60;
  const b = $('reenviar');
  clearInterval(espera);
  b.disabled = true;
  b.textContent = 'Reenviar código em ' + s + ' s';
  espera = setInterval(() => {
    s--;
    if (s <= 0) {
      clearInterval(espera);
      b.disabled = false;
      b.textContent = 'Reenviar código';
    } else {
      b.textContent = 'Reenviar código em ' + s + ' s';
    }
  }, 1000);
}
$('reenviar').addEventListener('click', () => pedirCodigo(emailCodigo, 'go-codigo', 'msg-codigo'));

// ---------- 3. código + senha nova ----------
$('codigo').addEventListener('input', (ev) => {
  const d = ev.target.value.replace(/\D/g, '').slice(0, 6);
  ev.target.value = d;
  if (d.length === 6) $('nova').focus();
});

$('form-codigo').addEventListener('submit', async (ev) => {
  ev.preventDefault();
  const codigo = $('codigo').value.replace(/\D/g, ''), senha = $('nova').value, senha2 = $('nova2').value;
  const senhaOk = senha.length >= 10 && /[A-Za-zÀ-ÿ]/.test(senha) && /[0-9]/.test(senha);
  erroCampo('fc', codigo.length !== 6);
  erroCampo('fn', !senhaOk);
  erroCampo('fn2', senhaOk && senha !== senha2);
  if (codigo.length !== 6) return aviso('msg-codigo', 'Digite os 6 dígitos do código que chegou no seu e-mail.');
  if (!senhaOk) return aviso('msg-codigo', 'A senha precisa ter pelo menos 10 caracteres, com letras e números.');
  if (senha !== senha2) return aviso('msg-codigo', 'As duas senhas estão diferentes.');
  aviso('msg-codigo', '');
  carregando('go-codigo', true);
  try {
    await api('POST', '/api/v1/auth/codigo/confirmar', { email: emailCodigo, codigo, senha });
    await irParaPlataforma();
  } catch (e) {
    carregando('go-codigo', false);
    erroCampo('fc', true);
    aviso('msg-codigo', e.message);
  }
});
