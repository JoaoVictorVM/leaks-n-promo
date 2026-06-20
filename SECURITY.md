# Política de Segurança

## Versões suportadas

O projeto está em desenvolvimento inicial (pré-1.0). Correções de segurança são
aplicadas apenas à versão mais recente da branch `main`.

| Versão | Suportada |
|---|---|
| `main` (mais recente) | ✅ |
| Versões anteriores | ❌ |

## Reportando uma vulnerabilidade

**Não** abra uma *issue* pública para relatar vulnerabilidades.

Prefira o canal privado do GitHub:
[**Report a vulnerability**](https://github.com/JoaoVictorVM/leaks-n-promo/security/advisories/new)
(*Security Advisories*). Alternativamente, envie um e-mail para
**jvmartinscv@gmail.com**.

Inclua, se possível:

- Descrição da vulnerabilidade e do impacto potencial.
- Passos para reproduzir (PoC).
- Versão/commit afetado.

Você receberá uma confirmação de recebimento e atualizações sobre o andamento da
correção. Pedimos um prazo razoável para correção antes de qualquer divulgação
pública (*coordinated disclosure*).

## Boas práticas adotadas no projeto

- **Stateless:** o bot não persiste dados de usuário (ver [PRIVACY.md](PRIVACY.md)).
- **Segredos** nunca são commitados (`.env` é gitignored; CI usa GitHub Secrets).
- **Scanners no CI:** `gitleaks`, `govulncheck`, CodeQL, Dependency Review, Trivy
  e OpenSSF Scorecard *(workflows a serem adicionados)*.
- **Supply chain:** Actions pinadas por commit SHA e SBOM gerado no release.
