// node --test robo-whatsapp/
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { pendentes, dadosProntos, ontemBR, mascarar, pausaAleatoria } from './regras.mjs';

test('só manda o que não foi enviado e tem telefone', () => {
  const l = [{ motorista: 'A', telefone: '5511900000000' }, { motorista: 'B', telefone: '' }, { motorista: 'C', telefone: '5511900000001', enviada: '2026-11-06T08:00:00Z' }];
  assert.deepEqual(pendentes(l).map((m) => m.motorista), ['A']);
  assert.deepEqual(pendentes(null), []);
});

test('espera os dados do dia chegarem', () => {
  assert.equal(dadosProntos({ situacao: 'ok', ultimo_dia: '2026-11-05' }, '2026-11-05'), true);
  assert.equal(dadosProntos({ situacao: 'ok', ultimo_dia: '2026-11-04' }, '2026-11-05'), false);
  assert.equal(dadosProntos({ situacao: 'rodando', ultimo_dia: '2026-11-05' }, '2026-11-05'), false);
  assert.equal(dadosProntos({ situacao: 'aguardando', ultimo_dia: '2026-11-06' }, '2026-11-05'), false);
  assert.equal(dadosProntos(null, '2026-11-05'), false);
});

test('ontem pelo horário de Brasília', () => {
  assert.equal(ontemBR(new Date('2026-11-06T11:00:00Z')), '2026-11-05');   // 08:00 em Brasília
  assert.equal(ontemBR(new Date('2026-11-06T02:00:00Z')), '2026-11-04');   // 23:00 do dia 5 em Brasília
  assert.equal(ontemBR(new Date('2026-03-01T12:00:00Z')), '2026-02-28');
});

test('telefone mascarado no log', () => {
  assert.equal(mascarar('5511900001234'), '••••1234');
  assert.equal(mascarar(''), 'sem telefone');
});

test('pausa entre as mensagens', () => {
  assert.equal(pausaAleatoria(25, 60, () => 0), 25000);
  assert.equal(pausaAleatoria(25, 60, () => 1), 60000);
});
