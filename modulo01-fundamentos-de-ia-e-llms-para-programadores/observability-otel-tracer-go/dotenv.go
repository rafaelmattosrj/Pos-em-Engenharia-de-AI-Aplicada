package main

import (
	"bufio"
	"os"
	"strings"
)

// loadDotEnv carrega variáveis de um arquivo .env simples para o ambiente do
// processo, sem sobrescrever variáveis já definidas — equivalente ao
// Dotenv.configure().ignoreIfMissing().load() usado na versão Java.
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

// getEnv retorna a variável de ambiente ou o valor padrão se ausente/vazia —
// equivalente a dotenv.get(key, defaultValue) na versão Java.
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
