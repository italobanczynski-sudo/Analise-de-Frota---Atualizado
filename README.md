# Análise de Frota (Gobrax)

Painel estático (HTML + JS), sem servidor próprio. Roda em dois lugares, com o mesmo `index.html`:

| Onde | Dados (clientes e planilhas) |
|---|---|
| **Painel no Claude** (Artifact): https://claude.ai/artifact/JtQke6GEAqzz8ohdP653dS | Base compartilhada: todos que têm acesso veem os mesmos dados, ao vivo. |
| GitHub Pages | Só no navegador de cada pessoa (localStorage). A página mostra um aviso com o link do painel no Claude. |

## Painel no Claude
- As planilhas ficam na coleção `planilhas` do banco do Artifact: um documento por tipo (motoristas/veículos), cliente e dia, dividido em partes (`~p1`, `~p2`…) quando passa de 200 KB.
- Para publicar, o leitor de Excel é carregado do cdnjs em vez de `vendor/` (troque o `src` do script ao gerar a cópia publicada).
- Acesso pelo menu Compartilhar do Artifact: Colaborador (Contributor) anexa e apaga planilhas; Leitor (Viewer) só vê.

## Publicar no GitHub Pages
1. Envie estes arquivos na raiz do repositório.
2. Settings → Pages → Source: "Deploy from a branch" → Branch `main` / pasta `/ (root)`.

## Observações
- O repositório/Pages contém só o código; nenhuma planilha ou dado de cliente vai junto.
