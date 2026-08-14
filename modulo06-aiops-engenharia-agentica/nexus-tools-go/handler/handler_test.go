package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"nexus-tools/service"
)

func newTestMux() *http.ServeMux {
	h := &Handlers{}
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestCheckComplianceRules_RetornaRegrasDePolitica(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest(http.MethodPost, "/tools/check-compliance-rules", bytes.NewReader([]byte(`{"query":"nomenclatura de buckets"}`)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}
	var resp toolResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Result == "" {
		t.Error("esperava resultado nao vazio")
	}
}

func TestSuggestFix_RetornaRemediacaoParaOOMKilled(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest(http.MethodPost, "/tools/suggest-fix", bytes.NewReader([]byte(`{"issueType":"OOMKilled"}`)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var resp toolResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Result != "Increase 'resources.limits.memory' to 1Gi in the Deployment spec." {
		t.Errorf("resultado inesperado: %q", resp.Result)
	}
}

func TestExecuteTerraform_BloqueiaAcaoDestrutivaSemSenha(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest(http.MethodPost, "/tools/execute-terraform", bytes.NewReader([]byte(`{"command":"destroy producao"}`)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var resp toolResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Result == "" {
		t.Fatal("esperava resultado")
	}
	if !strings.Contains(resp.Result, "BLOCKED") {
		t.Errorf("esperava bloqueio, obteve %q", resp.Result)
	}
}

func TestWriteFile_SalvaConteudoNoDisco(t *testing.T) {
	t.Cleanup(func() { os.Remove("controller-test-output.tf") })
	mux := newTestMux()

	req := httptest.NewRequest(http.MethodPost, "/tools/write-file",
		bytes.NewReader([]byte(`{"content":"resource \"aws_s3_bucket\" \"x\" {}","filename":"controller-test-output.tf"}`)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var resp toolResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !strings.Contains(resp.Result, "saved successfully") {
		t.Errorf("resultado inesperado: %q", resp.Result)
	}
}

func TestConsultRunbook_UsaDataPathConfigurado(t *testing.T) {
	tempDir := t.TempDir()
	os.WriteFile(tempDir+"/runbook_db.md", []byte("# Runbook de teste"), 0o644)

	h := &Handlers{Runbook: service.RunbookService{DataPath: tempDir}}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/tools/consult-runbook", bytes.NewReader([]byte(`{"serviceName":"db"}`)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var resp toolResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Result != "# Runbook de teste" {
		t.Errorf("resultado inesperado: %q", resp.Result)
	}
}
