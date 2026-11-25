/* =====================================================================
   TELA · Regras e integração
   ===================================================================== */
function vSettings(C){
  const u=S.u;if(!['regras','integracao','diagnostico'].includes(u.dataTab))u.dataTab='regras';
  const tabs=tabsHtml([['regras','Como calculamos'],['integracao','Integração'],['diagnostico','Diagnóstico do dado']],u.dataTab,'data-tab');
  const head=pageH('Regras e integração','Como a plataforma calcula cada número e de onde vem cada dado');
  return head+tabs+({regras:setRules,integracao:setIntegration,diagnostico:setDiag}[u.dataTab]||setRules)(C);
}
/* Diagnóstico: quanto de cada campo a API entregou para cada cliente (últimos 90 dias). Só contagens, sem nomes:
   mostra de uma vez se um campo vazio no painel é falta de dado na origem ou falha de tela. */
function diagCliente(c){
  const desde=dIso(T0-89);const rs=(c.registros||[]).filter(r=>String(r.dia).slice(0,10)>=desde);const rod=rs.filter(r=>r.rodou&&(+r.km||0)>0);
  const pct=f=>rod.length?Math.round(rod.filter(f).length/rod.length*100):null;const tem=k=>r=>r[k]!=null&&r[k]!=='';
  const meses=[...new Set((c.premiacoes||[]).filter(p=>p.nota!=null).map(p=>p.mes))].sort();
  return{nome:c.nome,dias:new Set(rs.map(r=>String(r.dia).slice(0,10))).size,veic:new Set(rs.map(r=>r.placa)).size,rod:rod.length,
    litros:pct(r=>(+r.litros||0)>0),freadas:pct(tem('frenagens')),vel:pct(tem('velocidade_media')),verde:pct(tem('faixa_verde')),ifv:pct(tem('inicio_faixa_verde')),ocioso:pct(tem('motor_ocioso')),
    motorista:pct(r=>!!r.motorista),nota:pct(tem('nota')),vinculos:(c.motoristas||[]).filter(r=>String(r.dia).slice(0,10)>=desde).length,
    premMeses:meses,premMot:new Set((c.premiacoes||[]).filter(p=>p.nota!=null).map(p=>p.motorista)).size,odo:(c.odometros||[]).length,
    ultimo:rs.length?rs.map(r=>String(r.dia).slice(0,10)).sort().pop():null};
}
function setDiag(C){
  if(!REAL.on)return`<div class="card sec rise"><div class="empty"><b>Disponível só na plataforma</b>O diagnóstico lê os dados reais do servidor. Abra pelo /painel com login.</div></div>`;
  const ds=(REAL.raw||[]).map(diagCliente);
  const cel=v=>v==null?'<td class="r muted">—</td>':`<td class="r num" style="font-weight:700;color:${v>=80?'var(--ok-t)':v>=30?'var(--warn-t)':'var(--bad-t)'}">${v}%</td>`;
  const rows=ds.map(d=>`<tr><td><b>${esc(d.nome)}</b><div class="xs muted">${d.veic} veículos · ${d.dias} dias · último ${d.ultimo?fmt.dm(d.ultimo):'—'}</div></td><td class="r num">${d.rod}</td>${cel(d.litros)}${cel(d.freadas)}${cel(d.vel)}${cel(d.verde)}${cel(d.ifv)}${cel(d.ocioso)}${cel(d.motorista)}<td class="r num muted">${d.nota==null?'—':d.nota+'%'}</td><td class="r num">${d.vinculos}</td><td class="r">${d.premMeses.length?`<b class="num">${d.premMot}</b> <span class="xs muted">${d.premMeses.map(m=>MESES[+m.slice(5)-1]).join(', ')}</span>`:'<span class="badge bad">nenhuma</span>'}</td><td class="r num">${d.odo}</td></tr>`).join('');
  const cid=C.cid,lv=REAL.live[cid],pos=lv&&lv.veiculos?Object.values(lv.veiculos).filter(v=>v.lat!=null).length:null,ras=lv&&lv.rastro?Object.keys(lv.rastro).length:null;
  const env=REAL.msgs[cid],tel=C.fleet.vehicles.filter(v=>v.phone).length,pm=premissasDe(C.def);
  const kv=(l,v,ok)=>`<div class="kv"><span>${l}</span><b class="num" style="color:${ok?'var(--ok-t)':'var(--bad-t)'}">${v}</b></div>`;
  return`<div class="card sec rise">${cardH('O que a API Gobrax entregou para cada cliente','Últimos 90 dias. Porcentagem dos dias rodados (km > 0) em que o campo veio preenchido. Verde 80%+, amarelo 30% a 79%, vermelho abaixo de 30%. Só contagens: nenhum nome aparece aqui.',`<button class="btn ghost sm" data-act="print">${ico('download',14)} PDF</button>`)}
  <div class="scroll-x"><table class="tbl"><thead><tr><th>Cliente</th><th class="r">Dias rodados</th><th class="r">Litros</th><th class="r">Freadas</th><th class="r">Velocidade</th><th class="r">Faixa verde</th><th class="r">Início f. verde</th><th class="r">Motor ocioso</th><th class="r">Motorista</th><th class="r">Nota do dia</th><th class="r">Vínculos motorista</th><th class="r">Premiação (motoristas · meses)</th><th class="r">Odômetros</th></tr></thead><tbody>${rows}</tbody></table></div>
  <div class="xs muted" style="margin-top:10px">Faixa verde, motor ocioso e motorista vêm da performance (v1/vehicle-performance); vazios aqui indicam que a busca de performance desse cliente não trouxe dado. "Nota do dia" vazia é esperado: a API não manda nota diária, a Nota Gobrax vem da premiação.</div></div>
  <div class="grid g2 sec"><div class="card rise stack">${cardH('Cliente aberto: '+esc(C.def.name),'O que o painel carregou além do dia a dia')}
    ${kv('Posição agora (veículos com GPS)',pos==null?(REAL.busy['live'+cid]?'carregando…':REAL.liveErr&&REAL.liveErr[cid]?'erro: '+REAL.liveErr[cid]:'não carregado'):pos+' de '+C.fleet.vehicles.length,pos>0)}
    ${kv('Rastro de hoje (veículos)',ras==null?'—':ras,ras>0)}
    ${kv('Mensagens marcadas como enviadas (motoristas)',env?env.length:'—',env&&env.length)}
    ${kv('Telefones de condutores',tel,tel>0)}
    ${kv('Contrato cadastrado (valor por unidade)',pm.valorEquip>0?fmt.brl(pm.valorEquip,2):'não (usa R$ 159,90)',pm.valorEquip>0)}
    ${kv('Diesel do cliente',pm.preco>0?fmt.brl(pm.preco,2)+'/L':'não (usa o padrão)',pm.preco>0)}</div>
  <div class="card brand rise stack"><div class="swoosh"></div>${cardH('Como usar','')}<p class="sm" style="margin:0">Se uma tela tem campo vazio, olhe a coluna dele aqui. Vermelho: a API não está mandando esse dado para o cliente (token sem acesso, veículo sem a unidade eletrônica ou falha na busca do servidor). Verde e a tela vazia: falha do painel, mande o print desta tabela e da tela.</p></div></div>`;
}
function setRules(C){
  const R=ST.rules;const kv=(l,v)=>`<div class="kv"><span>${l}</span><b class="num">${v}</b></div>`;
  return`<div class="grid g3 sec"><div class="card rise stack">${cardH('Índice de saúde do cliente','Pesos de cada sinal e faixas de risco')}${kv('Uso da frota',R.wUso)}${kv('Engajamento',R.wEng)}${kv('Nota Gobrax',R.wNota)}${kv('Resultado de consumo',R.wCons)}${kv('Motoristas críticos',R.wCrit)}${kv('Faixa média a partir de',R.bandMedio)}${kv('Faixa alta a partir de',R.bandAlto)}<div class="xs muted">Sinal sem dado sai do cálculo e os pesos se redistribuem.</div></div>
  <div class="card rise stack">${cardH('Qualidade do dado','Registros fora destes limites ficam fora das médias')}${kv('Consumo mínimo plausível',fmt.n(R.minKmpl,1)+' km/l')}${kv('Consumo máximo plausível',fmt.n(R.maxKmpl,1)+' km/l')}${kv('Máximo de km por dia',fmt.n(R.maxKmDay,0)+' km')}${kv('Comparação mínima',R.minCmpPct+'% dos dias')}<div class="xs muted">Dia sem litros ou sem km fica de fora do consumo. Indicador que a API não trouxe (nota, faixa verde, motor ocioso) fica fora da média daquele indicador, não vira zero.</div></div>
  <div class="card brand rise stack"><div class="swoosh"></div>${cardH('Premissas','Usadas no estudo, nos ganhos e nas oportunidades')}${kv('Preço do diesel',fmt.brl(R.diesel,2)+'/L')}${kv('Valor por unidade',fmt.brl(R.defUnit,2)+'/mês')}${kv('Nota mínima para bônus (estimativa)',R.bonusMin)}<div class="xs muted">Nota Gobrax: Bom a partir de 70, Atenção de 40 a 69 e Crítico abaixo de 40.</div></div></div>`;
}
function setIntegration(C){
  const map=[['totalMileage','km','Quilometragem do dia'],['totalConsumption','l','Litros consumidos'],['totalBreaking','br','Freadas'],['totalBreakingOnHighSpeed','brh','Freadas em alta velocidade'],['averageSpeed','sp','Velocidade média'],['indicators.greenRange.percentage','gr','Faixa verde (%)'],['indicators.idle.percentage','id','Motor ocioso (%)'],['score','sc','Nota Gobrax']];
  const fontes=[['/api/v1/painel/dados','Dia a dia de cada veículo (último ano), motoristas vinculados, odômetro e premiação'],['/api/v1/clientes/:id/ao-vivo','Onde cada veículo está agora e o rastro de hoje (mapa e trajeto)'],['/api/v1/clientes/:id/mensagens/enviadas','Mensagens marcadas como enviadas e a nota em cada envio'],['/api/v1/clientes/:id/condutores','Telefone de cada condutor (botão WhatsApp)'],['/mapa/:z/:x/:y','Mapa-base do OpenStreetMap, servido pelo próprio servidor']];
  const status=REAL.on?`<div class="banner info" style="margin:0 0 14px"><span>${ico('check',16)} Ligado à plataforma: os números vêm da API Gobrax, buscados pelo servidor de hora em hora.</span><a class="b" href="/plataforma">Clientes e coleta</a></div>`:`<div class="banner" style="margin:0 0 14px"><span>${ico('alert',16)} Modo demonstração: esta cópia não está atrás do login da plataforma, então mostra dados simulados.</span></div>`;
  return`${status}<div class="grid g-main sec"><div class="card rise">${cardH('De onde vem cada dado','Todas as telas leem de um único adaptador (Gobrax.api em src/data.js).')}<div class="scroll-x"><table class="tbl"><thead><tr><th>Rota do servidor</th><th>O que alimenta</th></tr></thead><tbody>${fontes.map(f=>`<tr><td><code>${f[0]}</code></td><td class="muted">${f[1]}</td></tr>`).join('')}</tbody></table></div><p class="sm muted">Cada registro é um dia de um veículo. Todo o resto (médias, risco, estudo, ganhos e perdas) é calculado no navegador a partir desses registros, com as mesmas regras em todas as telas.</p></div>
  <div class="card rise">${cardH('Campos da API → registro','')}<div class="scroll-x"><table class="tbl"><thead><tr><th>Campo da API</th><th>Registro</th><th>Significado</th></tr></thead><tbody>${map.map(m=>`<tr><td><code>${m[0]}</code></td><td><code>${m[1]}</code></td><td class="muted">${m[2]}</td></tr>`).join('')}</tbody></table></div></div></div>`;
}
