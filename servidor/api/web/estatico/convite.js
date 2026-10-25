// Tela do link de convite: a pessoa cria a senha e já entra.
'use strict';

const $ = (id) => document.getElementById(id);

// o token vem depois do # no link: o navegador não o envia ao servidor nos pedidos de página
const token = decodeURIComponent(location.hash.replace(/^#/, ''));
// tira o token da barra de endereço e do histórico
if (token) history.replaceState(null, '', location.pathname);

const aviso = (texto) => { $('erro').textContent = texto || ''; };
const carregando = (sim) => { $('botao').classList.toggle('load', sim); $('botao').disabled = sim; };

if (token.length !== 43) {
  aviso('Link de convite incompleto. Abra o link inteiro que você recebeu, ou use "Primeiro acesso" na tela de entrada.');
  $('botao').disabled = true;
}

document.querySelectorAll('.eye').forEach((b) => b.addEventListener('click', () => {
  const p = $(b.dataset.alvo), oculto = p.type === 'password';
  p.type = oculto ? 'text' : 'password';
  b.setAttribute('aria-label', oculto ? 'Ocultar senha' : 'Mostrar senha');
}));

$('form-convite').addEventListener('submit', async (ev) => {
  ev.preventDefault();
  const senha = $('senha').value, senha2 = $('senha2').value;
  const senhaOk = senha.length >= 10 && /[A-Za-zÀ-ÿ]/.test(senha) && /[0-9]/.test(senha);
  $('fs').classList.toggle('err', !senhaOk);
  $('fs2').classList.toggle('err', senhaOk && senha !== senha2);
  if (!senhaOk) return aviso('A senha precisa ter pelo menos 10 caracteres, com letras e números.');
  if (senha !== senha2) return aviso('As duas senhas estão diferentes.');
  aviso('');
  carregando(true);
  try {
    await api('POST', '/api/v1/auth/convite', { token, senha });
    await irParaPlataforma();
  } catch (e) {
    carregando(false);
    aviso(e.message);
  }
});
