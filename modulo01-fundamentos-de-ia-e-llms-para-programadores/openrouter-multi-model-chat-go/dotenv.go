package main

import (
	"bufio"
	"os"
	"strings"
)

// loadDotEnv carrega variáveis de um arquivo .env simples (KEY=VALUE por
// linha, '#' inicia comentário) para as variáveis de ambiente do processo,
// sem sobrescrever variáveis já definidas. Equivale ao Dotenv.configure()
// .ignoreIfMissing().load() usado na versão Java — se o arquivo não existir,
// simplesmente não faz nada.
func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}
