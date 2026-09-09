package main

import (
	"bufio"
	"os"
	"strings"
)

// loadDotEnv le um arquivo .env simples (KEY=VALUE por linha) e define as
// variaveis de ambiente que ainda nao existem. Nunca sobrescreve uma env var
// ja definida no processo. Mesmo padrao reaproveitado de
// pdf-rag-knowledge-base-go/dotenv.go e embeddings-vector-search-go/dotenv.go
// (equivalente a dotenv-java carregado em PspRoutingApplication.main na
// versao Java).
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return defaultValue
}
