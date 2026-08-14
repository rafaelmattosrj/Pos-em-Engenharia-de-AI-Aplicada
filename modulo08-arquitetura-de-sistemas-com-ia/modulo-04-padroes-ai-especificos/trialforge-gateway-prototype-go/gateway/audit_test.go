package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditTrailRegistrarAppendOnly(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "audit-trail.jsonl")
	trilha := NewAuditTrail(caminho)

	if err := trilha.Registrar(map[string]interface{}{"id_requisicao": "req-1", "status_final": "aprovado"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if err := trilha.Registrar(map[string]interface{}{"id_requisicao": "req-2", "status_final": "rejeitado"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	conteudo, err := os.ReadFile(caminho)
	if err != nil {
		t.Fatalf("erro ao ler trilha: %v", err)
	}
	linhas := strings.Split(strings.TrimSpace(string(conteudo)), "\n")
	if len(linhas) != 2 {
		t.Fatalf("esperado 2 linhas na trilha, obtido %d", len(linhas))
	}

	var registro map[string]interface{}
	if err := json.Unmarshal([]byte(linhas[0]), &registro); err != nil {
		t.Fatalf("linha inválida: %v", err)
	}
	if registro["id_requisicao"] != "req-1" {
		t.Errorf("id_requisicao = %v, esperado req-1", registro["id_requisicao"])
	}
	if _, temTimestamp := registro["timestamp"]; !temTimestamp {
		t.Errorf("esperado campo timestamp gravado automaticamente")
	}
}

// escreverTrilhaDemo grava exatamente o roteiro de 5 requisições que o
// próprio main.go executa (rotina / paráfrase-cache-hit / síntese-csr /
// tema-diferente-esgota-agentic / protocolo-1a-iteracao), pra exercitar
// VerificarTrilhaAuditoria sem precisar rodar contra um Ollama real.
func escreverTrilhaDemo(t *testing.T, caminho string, modeloCaro string) {
	t.Helper()
	trilha := NewAuditTrail(caminho)

	registros := []map[string]interface{}{
		// #1 rotina: sem cache hit, sem gate
		{
			"id_requisicao": "req-1", "cache_hit": false, "modelo_usado": "gemma4:e2b",
			"indice_usado": "icf", "iteracoes_agentic": 1, "esgotou_agentic": false,
			"confianca_rag": 0.9, "gate_acionado": false, "aprovado": true, "status_final": "aprovado",
		},
		// #2 paráfrase: cache HIT
		{
			"id_requisicao": "req-2", "cache_hit": true, "similaridade_cache": 0.9,
			"status_final": "respondido_via_cache",
		},
		// #3 síntese de CSR: pendência registrada antes da decisão + gate sempre acionado
		{"id_requisicao": "req-3", "status_final": "aguardando_aprovacao", "motivo_gate": "síntese de CSR"},
		{
			"id_requisicao": "req-3", "cache_hit": false, "modelo_usado": modeloCaro,
			"indice_usado": "csr", "iteracoes_agentic": 1, "esgotou_agentic": false,
			"confianca_rag": 0.95, "gate_acionado": true, "aprovado": true, "status_final": "aprovado",
		},
		// #4 tema diferente: cache miss, esgota Agentic RAG, gate por confiança baixa
		{"id_requisicao": "req-4", "status_final": "aguardando_aprovacao", "motivo_gate": "confiança baixa"},
		{
			"id_requisicao": "req-4", "cache_hit": false, "modelo_usado": "gemma4:e2b",
			"indice_usado": "csr", "iteracoes_agentic": 3, "esgotou_agentic": true,
			"confianca_rag": 0.5, "gate_acionado": true, "aprovado": true, "status_final": "aprovado",
		},
		// #5 protocolo: roteou pro índice certo, confiante já na 1ª iteração, sem gate
		{
			"id_requisicao": "req-5", "cache_hit": false, "modelo_usado": "gemma4:e2b",
			"indice_usado": "protocolo", "iteracoes_agentic": 1, "esgotou_agentic": false,
			"confianca_rag": 0.85, "gate_acionado": false, "aprovado": true, "status_final": "aprovado",
		},
	}

	for _, r := range registros {
		if err := trilha.Registrar(r); err != nil {
			t.Fatalf("erro ao gravar registro de teste: %v", err)
		}
	}
}

func TestVerificarTrilhaAuditoriaCaminhoFeliz(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "audit-trail.jsonl")
	escreverTrilhaDemo(t, caminho, "gemma4:latest")

	if err := VerificarTrilhaAuditoria(caminho, "gemma4:latest", 0.7, nil); err != nil {
		t.Fatalf("esperado trilha válida, obtido erro: %v", err)
	}
}

func TestVerificarTrilhaAuditoriaFalhaQuandoIncompleta(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "audit-trail.jsonl")
	trilha := NewAuditTrail(caminho)
	if err := trilha.Registrar(map[string]interface{}{"id_requisicao": "req-1", "status_final": "aprovado"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if err := VerificarTrilhaAuditoria(caminho, "gemma4:latest", 0.7, nil); err == nil {
		t.Fatal("esperado erro com trilha incompleta (menos de 5 requisições concluídas)")
	}
}
