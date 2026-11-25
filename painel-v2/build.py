"""Monta o arquivo único dist/analise-de-frota.html a partir de src/ e vendor/.
Uso:  python build.py            (gera dist/analise-de-frota.html)
      python build.py saida.html (caminho customizado)"""
import os, sys

ROOT = os.path.dirname(os.path.abspath(__file__))
SRC = os.path.join(ROOT, 'src')
VENDOR = os.path.join(ROOT, 'vendor')


def ler(*partes):
    with open(os.path.join(*partes), encoding='utf-8') as f:
        return f.read()


css = ler(SRC, 'style.css')
ordem = ['data.js', 'ui.js', 'charts.js', 'maps.js', 'views_a.js', 'views_b.js', 'views_d.js', 'views_c.js', 'app.js']
js = '\n'.join(ler(SRC, f) for f in ordem)
js = "(function(){'use strict';\n" + js + "\nboot();\n})();"
lj = ler(VENDOR, 'leaflet.js').replace('</script>', '<\\/script>')
lc = ler(VENDOR, 'leaflet.css')
html = (ler(SRC, 'template.html')
        .replace('/*__LCSS__*/', lc)
        .replace('/*__LJS__*/', lj)
        .replace('/*__CSS__*/', css)
        .replace('/*__JS__*/', js.replace('</script>', '<\\/script>')))
out = sys.argv[1] if len(sys.argv) > 1 else os.path.join(ROOT, 'dist', 'analise-de-frota.html')
os.makedirs(os.path.dirname(out), exist_ok=True)
with open(out, 'w', encoding='utf-8', newline='\n') as f:
    f.write(html)
print('OK', out, len(html), 'bytes')
