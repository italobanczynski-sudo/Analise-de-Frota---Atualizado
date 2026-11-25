/* =====================================================================
   DADOS · ADAPTADOR · AGREGAÇÕES
   Tudo que a tela mostra sai de UMA lista de registros diários por veículo.
   - No servidor (rota /painel, com login): os registros vêm de /api/v1/painel/dados,
     a mesma base que o servidor busca de hora em hora na API Gobrax
     (v1/vehicle-statistics + v1/vehicle-performance).
   - Fora do servidor (arquivo aberto direto, GitHub Pages): dados simulados.
   Formato do registro:
     { t:<dia em número>, date:'AAAA-MM-DD', veh:'PLACA',
       km  = totalMileage,  l  = totalConsumption (litros),
       br  = totalBreaking, brh = totalBreakingOnHighSpeed,
       sp  = averageSpeed,  gr = indicators.greenRange.percentage,
       id  = indicators.idle.percentage, sc = score (0–100) }
   Campos sem dado ficam null e saem das médias (não viram zero).
   ===================================================================== */
const DAY=864e5;
const TODAY=(()=>{try{return new Intl.DateTimeFormat('en-CA',{timeZone:'America/Sao_Paulo',year:'numeric',month:'2-digit',day:'2-digit'}).format(new Date())}catch(e){return new Date().toISOString().slice(0,10)}})();
const dIdx=iso=>{const [y,m,d]=iso.split('-').map(Number);return Math.round(Date.UTC(y,m-1,d)/DAY)};
const dIso=t=>new Date(t*DAY).toISOString().slice(0,10);
const wdOf=t=>new Date(t*DAY).getUTCDay();
const T0=dIdx(TODAY);
const clamp=(v,a,b)=>Math.max(a,Math.min(b,v));
const mulberry=a=>()=>{a|=0;a=a+0x6D2B79F5|0;let t=Math.imul(a^a>>>15,1|a);t=t+Math.imul(t^t>>>7,61|t)^t;return((t^t>>>14)>>>0)/4294967296};
const delay=ms=>new Promise(r=>setTimeout(r,ms));
const isoMinus=n=>dIso(T0-n);

/* Clientes simulados: as datas são relativas a hoje para o protótipo sempre ter histórico */
const CLIENT_DEFS=[
  {id:'via',name:'Via Group',vehicles:33,start:isoMinus(29),seed:11,mood:'dip',unit:159.9,units:33,prefixes:['SFI','TXM','TYI','SDY'],anom:14,silent:4},
  {id:'caju',name:'Caju Transportes',vehicles:42,start:isoMinus(62),seed:23,mood:'decline',unit:159.9,units:42,prefixes:['QAB','QBC','QCD','QDE'],anom:5,silent:3},
  {id:'xtsul',name:'Xtsul Transportes LTDA',vehicles:15,start:isoMinus(124),seed:37,mood:'improve',unit:159.9,units:15,prefixes:['RLM','RMN','RNP'],anom:2,silent:1}
];
const FIRST=['Adalberto','Benedito','Caio','Davi','Edgar','Fabrício','Geraldo','Heitor','Ismael','Jair','Kleber','Leandro','Matheus','Nilton','Osvaldo','Plínio','Quirino','Renan','Silvio','Tadeu','Valdir','Wagner','Yuri','Altair','Bruno','Cléber','Denis','Elton','Flávio','Gilson'];
const LAST=['Almeida','Barros','Cardoso','Duarte','Esteves','Furtado','Guedes','Holanda','Jardim','Klein','Lacerda','Medeiros','Nogueira','Ornelas','Pacheco','Queiroz','Rezende','Siqueira','Teles','Uchoa','Vasques','Xavier','Zanetti','Amaral','Bastos','Caldas','Dantas','Fontes','Godoy','Lima'];
const FILIAIS=['Matriz','Filial Sul','Filial Norte'];

const DEFAULT_RULES={minKmpl:1.0,maxKmpl:7.0,maxKmDay:1200,minKmEval:30,wUso:28,wEng:22,wNota:22,wCons:17,wCrit:11,bandAlto:60,bandMedio:35,minCmpPct:50,bonusMin:70,bonusRate:0.02,diesel:7.26,defUnit:159.9};

function moodFn(mood,start){
  const peak=T0-9;
  if(mood==='dip')return t=>-0.12*Math.exp(-Math.pow((t-peak)/5.5,2));
  if(mood==='decline')return t=>-0.0016*(t-dIdx(start));
  return t=>0.0009*(t-dIdx(start));
}

function genClient(def){
  const rnd=mulberry(def.seed);const pick=a=>a[Math.floor(rnd()*a.length)];
  const letters='ABCDEFGHJKLMNPRSTUVXZ';const used=new Set();const vehicles=[];const recs=[];
  const mood=moodFn(def.mood,def.start);const tStart=dIdx(def.start);
  const lowN=Math.max(2,Math.round(def.vehicles*0.09));const lateN=2;
  for(let i=0;i<def.vehicles;i++){
    let p;do{p=pick(def.prefixes)+Math.floor(rnd()*10)+letters[Math.floor(rnd()*letters.length)]+(10+Math.floor(rnd()*90))}while(used.has(p));used.add(p);
    const skill=(rnd()-0.5)*2;
    const roll=rnd();
    const slope=roll<0.3?0.004+rnd()*0.014:roll<0.65?-(0.003+rnd()*0.01):(rnd()-0.5)*0.004;
    const v={id:p,driver:pick(FIRST)+' '+pick(LAST),filial:pick(FILIAIS),base:clamp(1.55+skill*0.14+(rnd()-0.5)*0.1,1.3,1.9),slope,kmDay:250+rnd()*210,
      gr:clamp(58+skill*12+(rnd()-0.5)*6,40,82),idl:clamp(9-skill*3+(rnd()-0.5)*3,3,16),br:clamp(19-skill*4+(rnd()-0.5)*4,9,30),sp:34+rnd()*10,
      from:tStart,to:T0,low:false,phone:''};
    if(i<lowN){v.low=true;v.to=tStart+Math.floor((T0-tStart)*0.4)}
    else if(i<lowN+lateN){v.from=T0-8-Math.floor(rnd()*3)}
    else if(i>=def.vehicles-def.silent){v.to=T0-(4+Math.floor(rnd()*4))}
    vehicles.push(v);
  }
  for(const v of vehicles){
    for(let t=v.from;t<=v.to;t++){
      const w=wdOf(t);const p=(w===0?0.18:w===6?0.42:0.95)*(v.low?0.35:1);
      if(rnd()>p)continue;
      const km=Math.round(v.kmDay*(0.55+rnd()*0.85)*(w===0||w===6?0.6:1));
      if(km<5)continue;
      const kmpl=clamp(v.base+v.slope*((t-tStart)/7)+mood(t)+(rnd()-0.5)*0.12,1.0,2.7);
      const br=v.br*(0.8+rnd()*0.4);const gr=clamp(v.gr*(0.92+rnd()*0.16)+mood(t)*40,20,92);const idl=clamp(v.idl*(0.8+rnd()*0.4)-mood(t)*20,1,25);
      const sc=clamp(Math.round(28+gr*0.55-idl*1.4-br*0.55+(kmpl-1.5)*30+(rnd()-0.5)*8),4,95);
      recs.push({t,date:dIso(t),veh:v.id,km,l:+(km/kmpl).toFixed(1),br:Math.round(km*br/100),brh:Math.round(km*(rnd()<0.3?rnd()*1.2:0)/100),sp:+(v.sp*(0.9+rnd()*0.2)).toFixed(1),gr:+gr.toFixed(1),id:+idl.toFixed(1),sc});
    }
  }
  const recent=recs.filter(r=>r.t>T0-30&&r.km>120);
  for(let i=0;i<def.anom&&recent.length;i++){const r=recent.splice(Math.floor(rnd()*recent.length),1)[0];r.l=+(r.l*(1.7+rnd()*0.3)).toFixed(1)}
  recs.sort((a,b)=>a.t-b.t);
  return{vehicles,recs};
}

/* ---------- Servidor (plataforma com login) ---------- */
const REAL={on:false,eu:null,raw:null,live:{},liveAt:{},msgs:{},phones:{},busy:{},prem:[]};
async function srv(caminho,metodo='GET',corpo){
  const op={method:metodo,credentials:'same-origin',headers:{'X-Requisicao':'plataforma'}};
  if(corpo!==undefined){op.headers['Content-Type']='application/json';op.body=JSON.stringify(corpo)}
  const r=await fetch(caminho,op);
  if(r.status===401){location.replace('/');throw new Error('sessão encerrada')}
  let j=null;try{j=await r.json()}catch(e){}
  if(!r.ok||!j||!j.ok)throw new Error((j&&j.erro)||('O servidor respondeu '+r.status));
  return j.dados;
}
/* liga o modo servidor quando a página está atrás do login da plataforma */
async function detectServer(){
  if(!/^https?:$/.test(location.protocol))return false;
  try{REAL.eu=await srv('/api/v1/auth/eu');REAL.on=true;return true}catch(e){return false}
}
const nv=v=>v==null||v===''||!isFinite(+v)?null:+v;
const nkey=s=>String(s||'').normalize('NFD').replace(/[̀-ͯ]/g,'').toLowerCase().replace(/\s+/g,' ').trim();

/* um cliente de /api/v1/painel/dados → {vehicles, recs, ...} no formato das telas */
function fleetFromServer(c){
  const vmap=new Map(),drv={},recs=[],lastComm={};
  /* a performance diária da API não traz nota (o servidor grava nota vazia): a Nota Gobrax real é a do mês de cada
     motorista, na premiação (v2/driversOverview). Dia sem nota usa a nota do mês do motorista que dirigia o veículo. */
  const premNota={};for(const p of c.premiacoes||[]){const n=nv(p.nota);if(n!=null&&n>=0&&n<=100&&p.mes&&p.motorista)premNota[nkey(p.motorista)+'|'+p.mes]=n}
  const seen=(p,t)=>{if(lastComm[p]==null||t>lastComm[p])lastComm[p]=t};
  for(const r of c.registros||[]){
    const p=String(r.placa||'').trim();if(!p)continue;const dia=String(r.dia).slice(0,10);const t=dIdx(dia);
    if(!vmap.has(p))vmap.set(p,{id:p,driver:'',filial:'Frota',phone:''});
    if(r.motorista){(drv[p]=drv[p]||{})[r.motorista]=(drv[p][r.motorista]||0)+1}
    const km=nv(r.km)||0,l=nv(r.litros)||0;
    if(km>0)seen(p,t);
    if(!r.rodou||km<=0||l<=0)continue;
    let scd=nv(r.nota);if(scd!=null&&(scd<0||scd>100))scd=null;
    const nota=scd!=null?scd:r.motorista?premNota[nkey(r.motorista)+'|'+dia.slice(0,7)]??null:null;
    recs.push({t,date:dia,veh:p,km,l,br:nv(r.frenagens),brh:nv(r.frenagens_alta),sp:nv(r.velocidade_media),gr:nv(r.faixa_verde),ifv:nv(r.inicio_faixa_verde),id:nv(r.motor_ocioso),sc:nota,scd});
  }
  /* odômetro: a última comunicação da unidade eletrônica conta como sinal */
  for(const o of c.odometros||[]){const p=String(o.placa||'').trim();const d=String(o.coletado||'').slice(0,10);if(p&&/^\d{4}-\d\d-\d\d$/.test(d)){if(!vmap.has(p))vmap.set(p,{id:p,driver:'',filial:'Frota',phone:''});seen(p,dIdx(d))}}
  /* motorista de cada veículo: quem mais o dirigiu; e o veículo principal de cada motorista */
  const drvPlate={},cnt={};
  for(const r of c.motoristas||[]){if(!r.motorista||!r.placa)continue;const k=r.motorista;(cnt[k]=cnt[k]||{})[r.placa]=(cnt[k][r.placa]||0)+1}
  for(const k in cnt)drvPlate[nkey(k)]=Object.entries(cnt[k]).sort((a,b)=>b[1]-a[1])[0][0];
  /* sem motorista vinculado na API: a placa ocupa o lugar do nome (como no painel clássico) */
  vmap.forEach(v=>{const d=drv[v.id];v.driver=d?Object.entries(d).sort((a,b)=>b[1]-a[1])[0][0]:v.id;v.noDriver=!d;if(d)for(const n in d)if(!drvPlate[nkey(n)])drvPlate[nkey(n)]=v.id});
  recs.sort((a,b)=>a.t-b.t);
  const vehicles=[...vmap.values()].sort((a,b)=>a.id.localeCompare(b.id));
  return{vehicles,recs,lastComm,drvPlate,prem:c.premiacoes||[],srvId:c.id};
}
function clientsFromServer(raw){
  return raw.map(c=>{const f=fleetFromServer(c);const first=f.recs.length?f.recs[0].date:null;
    return{id:String(c.id),srvId:c.id,name:c.nome,vehicles:f.vehicles.length,units:f.vehicles.length,unit:DEFAULT_RULES.defUnit,start:first,_fleet:f}}).filter(c=>c.start);
}

/* ---------- Adaptador ---------- */
const Gobrax={api:{
  async getClients(){
    if(REAL.on){REAL.raw=await srv('/api/v1/painel/dados');return clientsFromServer(REAL.raw)}
    await delay(60);return CLIENT_DEFS.map(c=>({id:c.id,name:c.name,units:c.units,unit:c.unit,start:c.start}));
  },
  async getFleet(clientId){
    const def=CLIENT_DEFS.find(c=>c.id===clientId);
    if(REAL.on)return def._fleet||{vehicles:[],recs:[],lastComm:{},drvPlate:{},prem:[]};
    await delay(280);return genClient(def);
  },
  /* posição agora e rastro de hoje (servidor) */
  async getLive(clientId){
    const def=CLIENT_DEFS.find(c=>c.id===clientId);if(!REAL.on||!def)return null;
    const placas=DB.fleets[clientId].vehicles.map(v=>v.id);const out={veiculos:{},rastro:{}};
    for(let i=0;i<placas.length;i+=150){
      const d=await srv('/api/v1/clientes/'+def.srvId+'/ao-vivo?placas='+encodeURIComponent(placas.slice(i,i+150).join(',')));
      (d.veiculos||[]).forEach(v=>out.veiculos[v.placa]=v);Object.assign(out.rastro,d.rastro||{});
    }
    return out;
  },
  /* mensagens enviadas a cada motorista e a nota em cada envio (servidor) */
  async getSent(clientId){const def=CLIENT_DEFS.find(c=>c.id===clientId);if(!REAL.on||!def)return null;return srv('/api/v1/clientes/'+def.srvId+'/mensagens/enviadas')},
  /* telefones dos condutores (servidor; só CS e administrador) */
  async getPhones(clientId){const def=CLIENT_DEFS.find(c=>c.id===clientId);if(!REAL.on||!def)return null;return srv('/api/v1/clientes/'+def.srvId+'/condutores')},
  /* premissas de cada cliente (valor por unidade, unidades, diesel): a mesma coleção do painel clássico */
  async getPremissas(){if(!REAL.on)return[];const l=await srv('/api/v1/base/premissas');return(l||[]).map(d=>({id:d.id,...(d.dados||{})}))},
  async savePremissas(id,dados){return srv('/api/v1/base/premissas/'+encodeURIComponent(id),'PUT',dados)}
}};

/* ---------- Premissas do cliente (contrato e diesel) ---------- */
/* mesmas regras do painel clássico: documento "pr-<nome>" na coleção premissas; nome comparado sem acento/pontuação/LTDA */
const ckeyCli=v=>nkey(v).replace(/[.,/\-]/g,' ').replace(/\b(ltda|eireli|epp|me|s a|sa)\b/g,' ').replace(/\s+/g,' ').trim();
const idDocCli=t=>String(t).toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g,'').replace(/[^a-z0-9]+/g,'-').slice(0,140);
function premissasDe(def){
  if(!def)return{};
  if(!REAL.on){const c=ST.contracts[def.id]||{};return{valorEquip:c.unit,qtd:c.units}}
  return REAL.prem.find(d=>ckeyCli(d.cliente)===ckeyCli(def.name))||{};
}
/* contrato que vale nas contas: o digitado; sem ele, R$ 159,90 por veículo com dado */
function contratoDe(def){const pm=premissasDe(def);const unit=pm.valorEquip>0?+pm.valorEquip:DEFAULT_RULES.defUnit,units=pm.qtd>0?+pm.qtd:(def.vehicles||def.units||1);return{unit,units,unitAuto:!(pm.valorEquip>0),unitsAuto:!(pm.qtd>0)}}
const dieselDe=def=>{const pm=premissasDe(def);return pm.preco>0?+pm.preco:ST.rules.diesel};
async function salvarContrato(def,unit,units){
  if(!REAL.on){ST.contracts[def.id]={unit:unit||undefined,units:units||undefined};persist();return}
  const at=premissasDe(def);const mud=[];const R=v=>v?fmt.brl(v,2):'—';
  if((at.valorEquip||null)!==(unit||null))mud.push('valor por equipamento '+R(at.valorEquip)+' → '+R(unit));
  if((at.qtd||null)!==(units||null))mud.push('equipamentos '+(at.qtd||'—')+' → '+(units||'—'));
  if(!mud.length)return false;
  const nova={...at,cliente:at.cliente||def.name,valorEquip:unit||null,qtd:units||null};delete nova.id;
  nova.historico=[{em:new Date().toISOString(),por:(REAL.eu&&REAL.eu.nome)||'',mudou:mud.join('; ')+' (pelo painel novo)'}].concat(at.historico||[]).slice(0,100);
  const id=at.id||'pr-'+idDocCli(def.name);await Gobrax.api.savePremissas(id,nova);
  const i=REAL.prem.findIndex(d=>d.id===id);const doc={id,...nova};if(i>=0)REAL.prem[i]=doc;else REAL.prem.push(doc);return true;
}

/* ---------- Banco em memória ---------- */
const DB={fleets:{},cache:new Map(),ver:0,
  async load(id){if(!this.fleets[id]){this.fleets[id]=await Gobrax.api.getFleet(id);this.fleets[id].byId=Object.fromEntries(this.fleets[id].vehicles.map(v=>[v.id,v]))}return this.fleets[id]},
  bump(){this.ver++;this.cache.clear()}
};
const flagId=(cid,r)=>cid+':'+r.veh+':'+r.t;
function isSusp(r,rules){const k=r.km/r.l;return k<rules.minKmpl||k>rules.maxKmpl||r.km>rules.maxKmDay}
function validRecs(cid,store){
  const k='v'+cid;if(DB.cache.has(k))return DB.cache.get(k);
  const f=DB.fleets[cid];const out=[];
  for(const r of f.recs){const fl=store.flags[flagId(cid,r)];
    if(fl&&fl.l){const r2={...r,l:fl.l};if(!isSusp(r2,store.rules))out.push(r2);continue}
    if(!isSusp(r,store.rules)||fl==='confirmed')out.push(r)}
  DB.cache.set(k,out);return out;
}
function suspRecs(cid,store){return DB.fleets[cid].recs.filter(r=>{if(!isSusp(r,store.rules))return false;const fl=store.flags[flagId(cid,r)];if(fl==='confirmed'||fl==='ignored')return false;if(fl&&fl.l&&!isSusp({...r,l:fl.l},store.rules))return false;return true})}
/* médias ponderadas pelos km; campo sem dado (null) fica fora da média daquele indicador */
function aggregate(rs){
  const o={km:0,l:0,br:0,brh:0,n:rs.length};const w={sp:[0,0],gr:[0,0],ifv:[0,0],id:[0,0],sc:[0,0],scd:[0,0],br:[0,0],brh:[0,0]};
  for(const r of rs){o.km+=r.km;o.l+=r.l;
    for(const f in w){const v=r[f];if(v!=null&&isFinite(v)){w[f][0]+=(f==='br'||f==='brh')?v:v*r.km;w[f][1]+=r.km}}}
  const k=o.km;const wv=f=>w[f][1]?w[f][0]/w[f][1]:null;
  o.br=w.br[0];o.brh=w.brh[0];
  o.kmpl=o.l?o.km/o.l:null;o.kmDay=o.n?o.km/o.n:null;o.br100=w.br[1]?w.br[0]/w.br[1]*100:null;o.brh100=w.brh[1]?w.brh[0]/w.brh[1]*100:null;
  o.speed=wv('sp');o.green=wv('gr');o.ifv=wv('ifv');o.idle=wv('id');o.score=wv('sc');o.scd=wv('scd');
  o.lph100=k?o.l/k*100:null;
  return o;
}
function premFleet(prem,from,to){
  if(!REAL.on||!prem||!prem.length)return null;const ini=dIso(from).slice(0,7),fim=dIso(to).slice(0,7);
  /* média simples das notas dos motoristas, como a "Performance média" da tela de Premiação da plataforma Gobrax
     (cada motorista conta uma vez por mês, rodando muito ou pouco) */
  const calc=(a,b)=>{let s=0,n=0;const meses=new Set(),mots=new Set();
    for(const p of prem){if(!p.mes||p.mes<a||p.mes>b)continue;const nota=nv(p.nota);if(nota==null||nota<0||nota>100)continue;s+=nota;n++;meses.add(p.mes);mots.add(nkey(p.motorista))}
    return n?{nota:s/n,motoristas:mots.size,meses:[...meses].sort()}:null};
  const r=calc(ini,fim);if(r)return r;
  /* mês ainda sem premiação (o atual costuma sair depois): vale o último mês com premiação, até 2 meses antes */
  const ant=prem.map(p=>p.mes).filter(m=>m&&m<ini&&m>=prevMonth(prevMonth(ini))).sort().pop();
  const x=ant?calc(ant,ant):null;return x?{...x,stale:true}:null;
}
function metrics(cid,store,from,to,vehFilter){
  const key=['m',cid,from,to,vehFilter||'',DB.ver].join('|');if(DB.cache.has(key))return DB.cache.get(key);
  let rs=validRecs(cid,store).filter(r=>r.t>=from&&r.t<=to);
  if(vehFilter)rs=rs.filter(r=>vehFilter(r.veh));
  const m=aggregate(rs);m.len=to-from+1;m.from=from;m.to=to;
  const by={};for(const r of rs)(by[r.veh]=by[r.veh]||[]).push(r);
  m.byVeh={};for(const id in by){const a=aggregate(by[id]);a.id=id;m.byVeh[id]=a}
  m.active=Object.keys(by).length;
  const days=new Set(rs.map(r=>r.t));m.dataDays=days.size;
  const f=DB.fleets[cid];m.total=f.vehicles.length;
  /* Nota Gobrax da frota = a mesma conta da tela Premiação: média das notas dos motoristas no(s) mês(es) do período,
     ponderada pelos km (quem não rodou fica fora). Vale mesmo quando a API não diz quem dirigia cada veículo. */
  const pf=premFleet(f.prem,from,to);m.prem=pf;if(pf)m.score=pf.nota;
  let last=f.lastComm;if(!last){last={};for(const r of f.recs)last[r.veh]=r.t}
  m.noSignal=Object.keys(last).filter(v=>last[v]<T0-3).map(v=>({id:v,days:T0-last[v]})).sort((a,b)=>b.days-a.days);
  DB.cache.set(key,m);return m;
}
function dailySeries(cid,store,from,to,vehFilter){
  const rs=validRecs(cid,store).filter(r=>r.t>=from&&r.t<=to&&(!vehFilter||vehFilter(r.veh)));const by={};
  for(const r of rs)(by[r.t]=by[r.t]||[]).push(r);const out=[];
  for(let t=from;t<=to;t++){const a=by[t]?aggregate(by[t]):null;out.push({t,date:dIso(t),a})}
  return out;
}
function weeklySeries(cid,store,from,to){
  const out=[];let t=from;let n=1;
  while(t<=to){const dow=(wdOf(t)+6)%7;const end=Math.min(to,t+(6-dow));const a=metrics(cid,store,t,end);
    out.push({n,from:t,to:end,partial:(dow!==0&&t===from)||end-t<6,a});t=end+1;n++}
  return out;
}

/* ---------- Períodos ---------- */
function resolvePeriod(st){
  const p=st.period;let from,to=T0;
  if(p.preset==='7d')from=T0-6;else if(p.preset==='30d')from=T0-29;else if(p.preset==='90d')from=T0-89;
  else if(p.preset==='custom'){from=dIdx(p.from);to=dIdx(p.to)}
  else{const d=new Date(T0*DAY);from=dIdx(d.toISOString().slice(0,8)+'01')}
  const len=to-from+1;let cf,ct,cl;
  if(st.compare==='mes'){const d=new Date(from*DAY);const pm=new Date(Date.UTC(d.getUTCFullYear(),d.getUTCMonth()-1,1));cf=Math.round(pm/DAY);ct=from-1;cl='mês anterior ('+fmt.dm(dIso(cf))+' a '+fmt.dm(dIso(ct))+')'}
  else{ct=from-1;cf=ct-len+1;cl='período anterior ('+fmt.dm(dIso(cf))+' a '+fmt.dm(dIso(ct))+')'}
  return{from,to,len,cf,ct,clen:ct-cf+1,label:fmt.dm(dIso(from))+' a '+fmt.dm(dIso(to)),cmpLabel:cl};
}

/* ---------- Índice de risco (regras da plataforma) ---------- */
function health(cur,prev,rules){
  const S=[];const pct=(c,p)=>p?(c-p)/p*100:null;
  const cd=cur.km/cur.len,pd=prev.km/prev.len;
  const short=prev.dataDays<prev.len*rules.minCmpPct/100;const shortTxt='Base de comparação curta ('+prev.dataDays+' de '+prev.len+' dias com dado). Sinal fora do índice';
  if(short){S.push({k:'uso',label:'Uso da frota',w:rules.wUso,risk:null,txt:shortTxt});S.push({k:'eng',label:'Engajamento dos motoristas',w:rules.wEng,risk:null,txt:shortTxt})}
  else{
  if(prev.km>0&&cur.km>=0){const ch=pct(cd,pd);S.push({k:'uso',label:'Uso da frota',w:rules.wUso,risk:ch>=0?0:clamp(-ch*3,0,100),txt:'Km por dia do calendário '+fmt.sp(ch)+' vs. período anterior'})}else S.push({k:'uso',label:'Uso da frota',w:rules.wUso,risk:null,txt:'Sem base de comparação'});
  if(prev.n>0){const a=pct(cur.n/cur.len,prev.n/prev.len),b=pct(cur.active,prev.active);const part=x=>x<0?clamp(-x*2.8,0,100):0;S.push({k:'eng',label:'Engajamento dos motoristas',w:rules.wEng,risk:clamp((part(a)+part(b))/2,0,100),txt:cur.active+' veículos ativos (antes '+prev.active+') · '+fmt.sp(a)+' em dias trabalhados por dia'})}else S.push({k:'eng',label:'Engajamento dos motoristas',w:rules.wEng,risk:null,txt:'Sem base de comparação'});
  }
  if(cur.score!=null){const lv=clamp((70-cur.score)*2.5,0,100);let r=lv,t='Nota '+fmt.n(cur.score,1);if(prev.score!=null){const dr=Math.max(0,prev.score-cur.score);r=0.6*lv+0.4*clamp(dr*8,0,100);t+=' (antes '+fmt.n(prev.score,1)+')'}S.push({k:'nota',label:'Nota Gobrax',w:rules.wNota,risk:r,txt:t})}else S.push({k:'nota',label:'Nota Gobrax',w:rules.wNota,risk:null,txt:'Sem nota no período'});
  if(prev.kmpl&&cur.kmpl){const worse=Math.max(0,(prev.kmpl-cur.kmpl)/prev.kmpl*100);S.push({k:'cons',label:'Resultado de consumo',w:rules.wCons,risk:clamp(worse*5,0,100),txt:'Média '+fmt.n(cur.kmpl,2)+' km/l ('+fmt.sp(pct(cur.kmpl,prev.kmpl))+')'})}else S.push({k:'cons',label:'Resultado de consumo',w:rules.wCons,risk:null,txt:'Sem base de comparação'});
  const vs=Object.values(cur.byVeh).filter(v=>v.score!=null);
  if(vs.length){const sh=vs.filter(v=>v.score<40).length/vs.length*100;S.push({k:'crit',label:'Motoristas com nota crítica',w:rules.wCrit,risk:clamp(sh*2.5,0,100),txt:fmt.n(sh,0)+'% dos veículos com nota abaixo de 40'})}else S.push({k:'crit',label:'Motoristas com nota crítica',w:rules.wCrit,risk:null,txt:'Sem dado'});
  const av=S.filter(s=>s.risk!=null);const tw=av.reduce((a,s)=>a+s.w,0)||1;
  for(const s of S){s.share=s.risk!=null?s.w/tw*100:null;s.pts=s.risk!=null?s.risk*s.w/tw:null}
  const idx=av.reduce((a,s)=>a+s.pts,0);
  const band=idx>=rules.bandAlto?'ALTO':idx>=rules.bandMedio?'MÉDIO':'BAIXO';
  const top=[...av].sort((a,b)=>b.pts-a.pts)[0];
  return{idx:Math.round(idx),band,signals:S,top};
}

/* ---------- Oportunidades / ganhos e perdas ---------- */
function quant(arr,q){const a=arr.filter(x=>x!=null&&isFinite(x)).sort((x,y)=>x-y);if(!a.length)return null;const p=(a.length-1)*q;const lo=Math.floor(p),hi=Math.ceil(p);return a[lo]+(a[hi]-a[lo])*(p-lo)}
function opportunities(cur,prev,price){
  const vs=Object.values(cur.byVeh).filter(v=>v.km>150&&v.kmpl);
  const ref=quant(vs.map(v=>v.kmpl),0.75);
  const worse=ref?vs.filter(v=>v.kmpl<ref).map(v=>({id:v.id,kmpl:v.kmpl,money:(v.l-v.km/ref)*price})).sort((a,b)=>b.money-a.money):[];
  const potential=worse.reduce((a,v)=>a+v.money,0);
  const fell=[];for(const v of vs){const p=prev.byVeh[v.id];if(p&&p.kmpl&&p.km>150){const ch=(v.kmpl-p.kmpl)/p.kmpl*100;if(ch<-2)fell.push({id:v.id,from:p.kmpl,to:v.kmpl,ch})}}
  fell.sort((a,b)=>a.ch-b.ch);
  const kdRef=quant(vs.map(v=>v.kmDay),0.75);const low=kdRef?vs.filter(v=>v.kmDay<kdRef*0.8).sort((a,b)=>a.kmDay-b.kmDay):[];
  const brRef=quant(vs.map(v=>v.br100),0.25);const brW=brRef!=null?vs.filter(v=>v.br100!=null&&v.br100>brRef).length:0;
  return{ref,worse,potential,fell,evaluated:vs.length,kdRef,low,brRef,brW};
}
function gainsLosses(cur,prev,price){
  const out=[];
  for(const id in cur.byVeh){const v=cur.byVeh[id],p=prev.byVeh[id];if(!p||!p.kmpl||!v.kmpl||v.km<100)continue;const saved=(v.km/p.kmpl-v.l);out.push({id,money:saved*price,liters:saved,from:p.kmpl,to:v.kmpl})}
  out.sort((a,b)=>b.money-a.money);return out;
}
function vehStatus(cur,prev,id){
  const v=cur.byVeh[id],p=prev.byVeh[id];if(!v||!p||!p.kmpl||!v.kmpl)return{k:'base',label:'SEM BASE',cls:'n',ch:null};
  const ch=(v.kmpl-p.kmpl)/p.kmpl*100;if(ch>2)return{k:'up',label:'EVOLUIU',cls:'ok',ch};if(ch<-2)return{k:'down',label:'REGREDIU',cls:'bad',ch};return{k:'flat',label:'ESTÁVEL',cls:'n',ch};
}

/* ---------- Estudo do projeto ---------- */
function study(cid,store,price){
  const def=CLIENT_DEFS.find(c=>c.id===cid);const ct=contratoDe(def);
  const ps=dIdx(def.start);const len=T0-ps+1;const half=ps+Math.floor(len/2)-1;
  /* linha de base igual à do painel clássico: com 8 semanas ou mais, início = primeiras 4 semanas e agora = últimas 4;
     com menos, a primeira e a segunda metade do período */
  const nWk=weeklySeries(cid,store,ps,T0).length,curto=nWk<8;
  const pIni=curto?[ps,half]:[ps,ps+27],pAgora=curto?[half+1,T0]:[T0-27,T0];
  const rotIni=curto?'primeira metade do período':'primeiras 4 semanas',rotAgora=curto?'segunda metade':'últimas 4 semanas';
  const first=metrics(cid,store,pIni[0],pIni[1]);const base=first.kmpl||metrics(cid,store,ps,T0).kmpl||1;
  const mens=ct.unit*ct.units;const months=[];
  let cum=0,cumI=0;const weeks=[];
  const recs=validRecs(cid,store).filter(r=>r.t>=ps);
  const monthsMap={};
  for(const r of recs){const ym=r.date.slice(0,7);const m=monthsMap[ym]=monthsMap[ym]||{ym,km:0,l:0,sc:0,sk:0,days:new Set()};m.km+=r.km;m.l+=r.l;if(r.sc!=null){m.sc+=r.sc*r.km;m.sk+=r.km}m.days.add(r.t)}
  /* investimento por dia = mensalidade ÷ dias daquele mês (como no clássico): a soma dos dias dá o mês */
  const dim=t=>{const d=new Date(t*DAY);return new Date(Date.UTC(d.getUTCFullYear(),d.getUTCMonth()+1,0)).getUTCDate()};const invDia=t=>mens/dim(t);
  for(let t=ps;t<=T0;t++){const ym=dIso(t).slice(0,7);const m=monthsMap[ym]=monthsMap[ym]||{ym,km:0,l:0,sc:0,sk:0,days:new Set()};m.cal=(m.cal||0)+1;m.invc=(m.invc||0)+invDia(t)}
  for(const ym of Object.keys(monthsMap).sort()){const m=monthsMap[ym];const econ=(m.km/base-m.l)*price;const inv=m.invc||0;
    const [y,mo]=ym.split('-');const mi=dIdx(ym+'-01'),pfm=REAL.on?premFleet(DB.fleets[cid].prem,mi,mi):null;
    months.push({ym,label:MESES[+mo-1]+' '+y,from:fmt.dm(dIso(Math.max(ps,mi))),econ,inv,kmpl:m.l?m.km/m.l:null,score:pfm&&!pfm.stale?pfm.nota:m.sk?m.sc/m.sk:null,km:m.km,cal:m.cal})}
  const wk=weeklySeries(cid,store,ps,T0);
  for(const w of wk){const a=w.a;const e=a.km?(a.km/base-a.l)*price:0;let iv=0;for(let t=w.from;t<=w.to;t++)iv+=invDia(t);cum+=e;cumI+=iv;
    weeks.push({n:w.n,label:fmt.dm(dIso(w.to)),from:w.from,to:w.to,econ:e,inv:iv,cum,cumI,km:a.km,kmpl:a.kmpl,score:a.score,cal:w.to-w.from+1})}
  /* dia a dia: economia do dia contra a mesma linha de base, e o investimento rateado do dia */
  const days=dailySeries(cid,store,ps,T0).map(d=>{const a=d.a;return{t:d.t,date:d.date,econ:a?(a.km/base-a.l)*price:0,inv:invDia(d.t),km:a?a.km:0,kmpl:a?a.kmpl:null,score:a?a.score:null,n:a?a.n:0}});
  const econ=months.reduce((a,m)=>a+m.econ,0),inv=months.reduce((a,m)=>a+m.inv,0);
  const second=metrics(cid,store,pAgora[0],pAgora[1]);const pace=second.km?((second.km/base-second.l)*price)/(second.len)*30:0;
  const all=metrics(cid,store,ps,T0);
  return{ct,mens,base,first,second,months,weeks,days,econ,inv,miss:Math.max(0,inv-econ),roi:inv?econ/inv:null,pays:econ>=inv,pace,ps,half,len,all,pIni,pAgora,rotIni,rotAgora,evolved:null};
}
const MESES=['jan','fev','mar','abr','mai','jun','jul','ago','set','out','nov','dez'];

/* ---------- Premiação ---------- */
const prevMonth=ym=>{const [y,m]=ym.split('-').map(Number);const d=new Date(Date.UTC(y,m-2,1));return d.toISOString().slice(0,7)};
function premiacao(cid,store,ym){
  const f=DB.fleets[cid];const pym=prevMonth(ym);
  /* no servidor: o resultado oficial da premiação Gobrax (v2/driversOverview) quando existe para o mês */
  if(REAL.on&&f.prem&&f.prem.some(p=>p.mes===ym)){
    const prevBy={};f.prem.filter(p=>p.mes===pym&&p.nota!=null).forEach(p=>prevBy[nkey(p.motorista)]=Math.round(p.nota));
    return f.prem.filter(p=>p.mes===ym&&p.nota!=null).map(p=>({id:f.drvPlate[nkey(p.motorista)]||p.motorista,name:p.motorista,score:Math.round(p.nota),prev:prevBy[nkey(p.motorista)]??null,km:nv(p.km)||0,bonus:nv(p.premio)||0,official:true})).sort((x,y)=>y.score-x.score);
  }
  const rs=validRecs(cid,store).filter(r=>r.date.slice(0,7)===ym);const by={};
  for(const r of rs)(by[r.veh]=by[r.veh]||[]).push(r);
  const prs=validRecs(cid,store).filter(r=>r.date.slice(0,7)===pym);const pb={};for(const r of prs)(pb[r.veh]=pb[r.veh]||[]).push(r);
  return Object.keys(by).map(id=>{const a=aggregate(by[id]);if(a.score==null)return null;const p=pb[id]?aggregate(pb[id]):null;const sc=Math.round(a.score);const bonus=sc>=store.rules.bonusMin?a.km*store.rules.bonusRate:0;
    return{id,name:f.byId[id].driver,score:sc,prev:p&&p.score!=null?Math.round(p.score):null,km:a.km,bonus}}).filter(Boolean).sort((x,y)=>y.score-x.score);
}
function premMonths(cid,store){const s=new Set(validRecs(cid,store).map(r=>r.date.slice(0,7)));const f=DB.fleets[cid];if(REAL.on&&f.prem)f.prem.forEach(p=>p.mes&&s.add(p.mes));return[...s].sort()}
function scoreBand(s){return s==null||!isFinite(s)?{k:'e',label:'SEM NOTA',cls:'n'}:s>=70?{k:'g',label:'BOM',cls:'ok'}:s>=40?{k:'y',label:'ATENÇÃO',cls:'warn'}:{k:'r',label:'CRÍTICO',cls:'bad'}}

/* ---------- Mensagem do dia ---------- */
function dayMessage(cid,store,veh,t){
  const f=DB.fleets[cid];const v=f.byId[veh];const recs=validRecs(cid,store).filter(r=>r.veh===veh&&r.t<=t);
  const cur=recs.find(r=>r.t===t);if(!cur)return null;
  const prevR=[...recs].reverse().find(r=>r.t<t);const m=dIso(t).slice(0,7);const month=aggregate(recs.filter(r=>r.date.slice(0,7)===m));
  const kmpl=cur.km/cur.l;const pk=prevR?prevR.km/prevR.l:null;
  const pts=[];const ok=x=>x!=null&&isFinite(x);
  if(prevR){if(ok(cur.gr)&&ok(prevR.gr))pts.push({k:'Faixa verde',d:cur.gr-prevR.gr,txt:'aumente o tempo na faixa verde do conta-giros'});if(ok(cur.id)&&ok(prevR.id))pts.push({k:'Motor ocioso',d:-(cur.id-prevR.id),txt:'evite deixar o motor ligado parado nas esperas'});if(ok(cur.br)&&ok(prevR.br))pts.push({k:'Freadas',d:-(cur.br/cur.km*100-prevR.br/prevR.km*100),txt:'mantenha mais distância do veículo da frente para frear menos'})}
  const worst=pts.sort((a,b)=>a.d-b.d)[0];
  const first=v.noDriver?'':v.driver.split(' ')[0];const score=month.score!=null?Math.round(month.score):null;
  let txt='Olá'+(first?', '+first:'')+'! Aqui é a Gobrax. Seu dia '+fmt.dm(dIso(t))+': '+fmt.n(cur.km,0)+' km rodados e média de '+fmt.n(kmpl,2)+' km/l';
  if(pk)txt+=' ('+fmt.sp((kmpl-pk)/pk*100)+' contra o último dia em que você rodou)';
  txt+='.';if(score!=null)txt+=' Sua nota no mês é '+score+' ('+scoreBand(score).label.toLowerCase()+').';
  txt+=worst?' Para melhorar: '+worst.txt+'. Conta com a gente!':' Continue assim!';
  return{veh,name:v.driver,first,km:cur.km,kmpl,pk,point:worst?worst.k:null,pointD:worst?worst.d:null,score,text:txt,phone:v.phone};
}
