# Política de Privacidade

O **leaks&promo** foi projetado para ser **stateless** na v1: ele não mantém
banco de dados nem armazena informações sobre quem usa os comandos.

## O que o bot **não** armazena

- ❌ Identificadores de usuário do Discord.
- ❌ Histórico de comandos ou termos pesquisados.
- ❌ Conteúdo de mensagens.
- ❌ Qualquer dado pessoal (PII).

## O que acontece com as suas consultas

- Quando você usa `/preco` ou `/leaks`, o termo é utilizado **apenas em memória**
  para realizar a busca nas fontes externas e montar a resposta.
- Pode haver um **cache em memória, com TTL curto**, para reduzir chamadas às
  APIs externas e melhorar a latência. Esse cache é volátil e some quando o
  processo é reiniciado.

## Logs

- Os logs são estruturados e **não** incluem PII nem segredos.

## Fontes externas

Ao responder, o bot consulta serviços de terceiros (CheapShark, fontes RSS e,
opcionalmente, Reddit). O uso desses serviços está sujeito às respectivas
políticas. As atribuições estão documentadas em [`NOTICE`](NOTICE).

## Futuro

Uma eventual funcionalidade de **notificações push** (Fase 3 / v2.0) poderá
exigir persistência mínima. Se e quando isso acontecer, esta política será
atualizada antes do recurso entrar em produção.
