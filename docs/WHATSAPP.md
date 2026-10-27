# WhatsApp dos motoristas: como ligar

Tudo é feito na plataforma, em **WhatsApp dos motoristas** (botão com as suas iniciais no painel, ou
https://analise-de-frota.onrender.com/plataforma#whatsapp). Só administrador vê a parte da Twilio.

> Account SID, Auth Token e telefones vão **só** na plataforma. Nunca em chat, e-mail, print ou planilha compartilhada
> (Política de Segurança da Informação da Gobrax, Cláusula 5.1).

## 1. Preparar a Twilio (uma vez)

Em console.twilio.com:

1. **Sandbox, para testar:** *Messaging* → *Try it out* → *Send a WhatsApp message*.
   Do seu WhatsApp, mande o `join ...` que aparece na tela para o número do Sandbox.
2. **Template:** *Messaging* → *Content Template Builder* → *Create new* → tipo **Text**, idioma **Portuguese (Brazil)**.
   No corpo, cole o texto do botão **copiar texto do template** (fica na Conexão com a Twilio, na plataforma).
   Salve e copie o **Content SID** (começa com `HX`).

## 2. Preencher a Conexão com a Twilio

Na aba WhatsApp, cartão **Conexão com a Twilio**:

| Campo | O que colocar |
|---|---|
| Account SID | Console da Twilio → página inicial → *Account Info* (começa com `AC`) |
| Auth Token | Mesma tela → *Show* |
| Número que envia | Já vem com o do Sandbox (`+1 415 523 8886`) |
| Template SID | O `HX...` do passo 1 |
| Enviar sozinho todo dia | Deixe **desmarcado** até testar |

Clique em **Salvar conexão**. O aviso no topo muda para **"Twilio conectada"**.

## 3. Testar com o seu número

1. Escolha o cliente.
2. Em **Cadastrar WhatsApp**, escolha um motorista que rodou ontem, coloque **o seu número** e salve.
3. **Ver prévia** → **Enviar agora**. A mensagem chega no seu WhatsApp.
4. Depois, salve de novo o mesmo motorista com o número dele.

## 4. Ligar para os motoristas

Para os motoristas receberem, é preciso um número WhatsApp Business na Twilio e o template aprovado pelo WhatsApp
(categoria **Utility**). Com os dois prontos, troque o **Número que envia**, marque **Enviar sozinho todo dia**,
escolha a hora e salve. Ninguém recebe o mesmo dia duas vezes.

## Se der erro

| O que aparece | O que fazer |
|---|---|
| "Account SID inválido", "Auth Token inválido"... | Copie de novo do console, sem espaços |
| "A Twilio recusou o Account SID e o Auth Token" | Token trocado ou errado: cole o atual e salve |
| Erro **63015** | O seu número não mandou o `join ...` para o Sandbox |
| Erro **63016** | Template SID errado ou template não aprovado |

Depois dos testes, troque o Auth Token na Twilio (*Rotate*) e salve o novo na Conexão com a Twilio.
