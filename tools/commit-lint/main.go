package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	message, err := readMessage(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "commit-lint:", err)
		os.Exit(2)
	}

	problems := Validate(message)
	if len(problems) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "Mensagem de commit inválida (Conventional Commits):")
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, "  -", p)
	}
	fmt.Fprintln(os.Stderr, "\nFormato esperado: tipo(escopo opcional): descrição no imperativo, sem ponto final")
	os.Exit(1)
}

func readMessage(args []string) (string, error) {
	if len(args) > 0 {
		data, err := os.ReadFile(args[0]) //nolint:gosec // caminho controlado pelo git/CI
		if err != nil {
			return "", fmt.Errorf("lendo arquivo de mensagem %q: %w", args[0], err)
		}
		return string(data), nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("lendo mensagem de stdin: %w", err)
	}
	return string(data), nil
}
