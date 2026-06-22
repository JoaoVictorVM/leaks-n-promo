package main

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		wantValid bool
		// wantContains é um trecho esperado em algum problema (só quando inválido).
		wantContains string
	}{
		{
			name:      "tipo simples",
			message:   "feat: add price command",
			wantValid: true,
		},
		{
			name:      "com escopo",
			message:   "fix(price): handle 429 from cheapshark",
			wantValid: true,
		},
		{
			name:      "breaking change com bang",
			message:   "feat(api)!: drop v1 endpoint",
			wantValid: true,
		},
		{
			name:      "tipo revert",
			message:   "revert: undo cache change",
			wantValid: true,
		},
		{
			name:      "corpo com linha em branco",
			message:   "feat: add cache\n\ncorpo explicando a decisão",
			wantValid: true,
		},
		{
			name:      "remove BOM antes de validar",
			message:   bom + "feat: add price command",
			wantValid: true,
		},
		{
			name:         "tipo nao permitido",
			message:      "feet: add thing",
			wantValid:    false,
			wantContains: "não é permitido",
		},
		{
			name:         "tipo maiusculo nao casa formato",
			message:      "Feat: add thing",
			wantValid:    false,
			wantContains: "formato inválido",
		},
		{
			name:         "sem espaco apos dois-pontos",
			message:      "feat:add thing",
			wantValid:    false,
			wantContains: "formato inválido",
		},
		{
			name:         "termina com ponto",
			message:      "feat: add thing.",
			wantValid:    false,
			wantContains: "ponto final",
		},
		{
			name:         "descricao nao imperativa",
			message:      "feat: adds the thing",
			wantValid:    false,
			wantContains: "imperativo",
		},
		{
			name:         "cabecalho muito longo",
			message:      "feat: " + strings.Repeat("x", maxHeaderLength),
			wantValid:    false,
			wantContains: "máximo",
		},
		{
			name:         "mensagem vazia",
			message:      "",
			wantValid:    false,
			wantContains: "vazia",
		},
		{
			name:         "corpo sem linha em branco",
			message:      "feat: add cache\ncorpo colado no cabeçalho",
			wantValid:    false,
			wantContains: "linha em branco",
		},
		{
			name:      "ignora merge commit",
			message:   "Merge branch 'main' into feature",
			wantValid: true,
		},
		{
			name:      "ignora revert padrao do git",
			message:   `Revert "feat: add price command"`,
			wantValid: true,
		},
		{
			name:      "ignora fixup",
			message:   "fixup! feat: add price command",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := Validate(tt.message)

			if tt.wantValid {
				if len(problems) != 0 {
					t.Fatalf("esperava mensagem válida, mas obtive problemas: %v", problems)
				}
				return
			}

			if len(problems) == 0 {
				t.Fatalf("esperava problemas, mas a mensagem foi considerada válida")
			}
			if !containsSubstring(problems, tt.wantContains) {
				t.Fatalf("nenhum problema contém %q; problemas: %v", tt.wantContains, problems)
			}
		})
	}
}

func TestValidateAccumulatesProblems(t *testing.T) {
	// Tipo inválido + ponto final + não-imperativo devem ser reportados juntos.
	problems := Validate("nope: adds the thing.")
	if len(problems) < 2 {
		t.Fatalf("esperava múltiplos problemas, obtive: %v", problems)
	}
}

func TestFirstWord(t *testing.T) {
	tests := map[string]string{
		"add price command": "add",
		"add":               "add",
		"  trim  ":          "trim",
		"":                  "",
	}
	for in, want := range tests {
		if got := firstWord(in); got != want {
			t.Errorf("firstWord(%q) = %q, esperava %q", in, got, want)
		}
	}
}

func containsSubstring(problems []string, want string) bool {
	for _, p := range problems {
		if strings.Contains(p, want) {
			return true
		}
	}
	return false
}
