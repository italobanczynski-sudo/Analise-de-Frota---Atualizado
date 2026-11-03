// Confere a sintaxe de cada <script> inline do index.html
const fs = require('fs'), vm = require('vm');
const h = fs.readFileSync(process.argv[2], 'utf8');
const re = /<script>([\s\S]*?)<\/script>/g; let m, n = 0, err = 0;
while ((m = re.exec(h))) { n++; try { new vm.Script(m[1]); } catch (e) { err++; console.log('script', n, e.message); } }
console.log(n + ' scripts, ' + err + ' com erro');
