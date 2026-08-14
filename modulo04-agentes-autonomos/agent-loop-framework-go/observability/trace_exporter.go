package observability

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"agent-loop-framework/model"
)

// TraceExporter serializa o trace do agente para JSON para análise
// post-mortem — equivalente a TraceExporter.java.
type TraceExporter struct{}

// ExportToJSON serializa o AgentTrace para uma string JSON formatada
// (pretty-print).
func (TraceExporter) ExportToJSON(trace *model.AgentTrace) string {
	out, err := json.MarshalIndent(trace, "", "  ")
	if err != nil {
		log.Printf("[TraceExporter] Falha ao serializar trace: %v", err)
		return fmt.Sprintf(`{"error": "Falha ao serializar trace: %s"}`, err)
	}
	return string(out)
}

// SaveToFile salva o AgentTrace em um arquivo JSON no caminho especificado.
func (e TraceExporter) SaveToFile(trace *model.AgentTrace, filename string) {
	jsonText := e.ExportToJSON(trace)

	if dir := filepath.Dir(filename); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Printf("[TraceExporter] Falha ao criar diretorio para '%s': %v", filename, err)
			return
		}
	}

	if err := os.WriteFile(filename, []byte(jsonText), 0o644); err != nil {
		log.Printf("[TraceExporter] Falha ao salvar trace em '%s': %v", filename, err)
		return
	}
	log.Printf("[TraceExporter] Trace salvo em: %s", filename)
}
