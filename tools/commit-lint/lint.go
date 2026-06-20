package main

import (
	"fmt"
	"regexp"
	"strings"
)

const maxHeaderLength = 72

var bom = string(rune(0xFEFF))

var allowedTypes = []string{
	"feat", "fix", "docs", "chore", "ci",
	"build", "test", "refactor", "perf", "style", "revert",
}

var headerPattern = regexp.MustCompile(`^([a-z]+)(?:\(([^)]+)\))?(!)?: (.+)$`)

var nonImperative = map[string]bool{
	"added": true, "adds": true, "adding": true,
	"fixed": true, "fixes": true, "fixing": true,
	"updated": true, "updates": true, "updating": true,
	"removed": true, "removes": true, "removing": true,
	"changed": true, "changes": true, "changing": true,
	"created": true, "creates": true, "creating": true,
	"implemented": true, "implements": true, "implementing": true,
	"refactored": true, "refactors": true, "refactoring": true,
	"deleted": true, "deletes": true, "deleting": true,
}

var ignoredPrefixes = []string{"Merge ", `Revert "`, "fixup!", "squash!", "amend!"}

func Validate(message string) []string {
	message = strings.TrimPrefix(message, bom)

	header := headerLine(message)
	if header == "" {
		return []string{"a mensagem de commit está vazia"}
	}

	if isIgnored(header) {
		return nil
	}

	var problems []string

	if n := len(header); n > maxHeaderLength {
		problems = append(problems, fmt.Sprintf("o cabeçalho tem %d caracteres (máximo %d)", n, maxHeaderLength))
	}

	m := headerPattern.FindStringSubmatch(header)
	if m == nil {
		problems = append(problems, `formato inválido; use "tipo(escopo): descrição" com tipo em minúsculas e espaço após os dois-pontos`)
		return problems
	}

	typ, desc := m[1], m[4]

	if !isAllowedType(typ) {
		problems = append(problems, fmt.Sprintf("tipo %q não é permitido (use um de: %s)", typ, strings.Join(allowedTypes, ", ")))
	}

	if strings.HasSuffix(desc, ".") {
		problems = append(problems, "a descrição não deve terminar com ponto final")
	}

	if first := firstWord(desc); nonImperative[strings.ToLower(first)] {
		problems = append(problems, fmt.Sprintf("a descrição deve usar o imperativo (ex.: \"add\" no lugar de %q)", first))
	}

	if bodyMissingBlankLine(message) {
		problems = append(problems, "deve haver uma linha em branco entre o cabeçalho e o corpo")
	}

	return problems
}

func headerLine(message string) string {
	for _, line := range splitLines(message) {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if trimmed := strings.TrimRight(line, " \t"); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func bodyMissingBlankLine(message string) bool {
	lines := splitLines(message)

	i := 0
	for i < len(lines) && (strings.HasPrefix(lines[i], "#") || strings.TrimRight(lines[i], " \t") == "") {
		i++
	}

	next := i + 1
	if next >= len(lines) || strings.HasPrefix(lines[next], "#") {
		return false
	}
	return strings.TrimRight(lines[next], " \t") != ""
}

func isIgnored(header string) bool {
	for _, p := range ignoredPrefixes {
		if strings.HasPrefix(header, p) {
			return true
		}
	}
	return false
}

func isAllowedType(typ string) bool {
	for _, t := range allowedTypes {
		if t == typ {
			return true
		}
	}
	return false
}

func firstWord(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexAny(s, " \t"); idx >= 0 {
		return s[:idx]
	}
	return s
}

func splitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}
