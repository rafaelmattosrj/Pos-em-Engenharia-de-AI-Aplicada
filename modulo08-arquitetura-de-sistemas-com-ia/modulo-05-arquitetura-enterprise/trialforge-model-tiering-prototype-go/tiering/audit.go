package tiering

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// AuditTrail é a trilha de auditoria: registro append-only em JSON Lines, um
// registro por requisição processada — mesma disciplina do Módulo 4.5, relida
// depois para confirmar as DECISÕES determinísticas (tier usado, se escalou,
// se bloqueou por orçamento), nunca o texto exato gerado pelo modelo.
//
// Adaptação: cada porte (JS/Python original, Java, Go) grava em seu próprio
// arquivo local audit-trail-tiering.jsonl, dentro da respectiva pasta do
// projeto — não no arquivo compartilhado da pasta pai — para as execuções de
// diferentes linguagens não colidirem/disputarem o mesmo arquivo. Ver README.
type AuditTrail struct {
	caminho string
	mu      sync.Mutex
}

func NewAuditTrail(caminho string) *AuditTrail {
	return &AuditTrail{caminho: caminho}
}

// Registrar adiciona timestamp e grava o registro como uma linha JSON, em
// modo append — protegido por mutex para permitir chamadas concorrentes
// (ver SimularVolumeConcorrente).
func (a *AuditTrail) Registrar(registro map[string]any) error {
	linha := make(map[string]any, len(registro)+1)
	linha["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	for k, v := range registro {
		linha[k] = v
	}

	dados, err := json.Marshal(linha)
	if err != nil {
		return fmt.Errorf("audit trail: falha ao serializar registro: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	arquivo, err := os.OpenFile(a.caminho, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("audit trail: falha ao abrir %s: %w", a.caminho, err)
	}
	defer arquivo.Close()

	if _, err := arquivo.Write(append(dados, '\n')); err != nil {
		return fmt.Errorf("audit trail: falha ao gravar em %s: %w", a.caminho, err)
	}
	return nil
}

// LerTodas lê todas as linhas do arquivo, cada uma decodificada como um mapa
// genérico (equivalente a JsonNode/dict).
func (a *AuditTrail) LerTodas() ([]map[string]any, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	arquivo, err := os.Open(a.caminho)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("audit trail: falha ao abrir %s: %w", a.caminho, err)
	}
	defer arquivo.Close()

	var resultado []map[string]any
	scanner := bufio.NewScanner(arquivo)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())
		if linha == "" {
			continue
		}
		var registro map[string]any
		if err := json.Unmarshal([]byte(linha), &registro); err != nil {
			return nil, fmt.Errorf("audit trail: linha inválida: %w", err)
		}
		resultado = append(resultado, registro)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("audit trail: falha ao ler %s: %w", a.caminho, err)
	}
	return resultado, nil
}
