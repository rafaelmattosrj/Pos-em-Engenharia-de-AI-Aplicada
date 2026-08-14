package service

import (
	"os"
	"strings"
	"testing"
)

func cleanupFile(t *testing.T, filename string) {
	t.Helper()
	t.Cleanup(func() { os.Remove(filename) })
}

// --- K8sOpsService ---

func TestGenerateK8sManifest_EscreveArquivoComConteudoEsperado(t *testing.T) {
	cleanupFile(t, "checkout-k8s.yaml")
	svc := K8sOpsService{}

	result := svc.GenerateK8sManifest("checkout", 3, 8080)
	if !strings.Contains(result, "successfully generated") {
		t.Errorf("resultado inesperado: %q", result)
	}

	content, err := os.ReadFile("checkout-k8s.yaml")
	if err != nil {
		t.Fatalf("erro lendo arquivo gerado: %v", err)
	}
	for _, expected := range []string{"name: checkout", "replicas: 3", "containerPort: 8080"} {
		if !strings.Contains(string(content), expected) {
			t.Errorf("esperava manifesto contendo %q", expected)
		}
	}
}

func TestApplyK8sManifest_ArquivoInexistenteRetornaErro(t *testing.T) {
	svc := K8sOpsService{}
	result := svc.ApplyK8sManifest("arquivo-que-nao-existe.yaml")
	if !strings.Contains(result, "was not found to apply") {
		t.Errorf("resultado inesperado: %q", result)
	}
}

func TestAnalyzeCanaryMetrics_ErroDisparaRollback(t *testing.T) {
	svc := K8sOpsService{}
	if !strings.Contains(svc.AnalyzeCanaryMetrics("error_rate > 5%"), "ROLLBACK") {
		t.Error("esperava ROLLBACK")
	}
	if !strings.Contains(svc.AnalyzeCanaryMetrics("all metrics stable"), "PROCEED") {
		t.Error("esperava PROCEED")
	}
}

// --- K8sDiagService ---

func TestInspectPodFailure_DiferenciaPorNomeDoPod(t *testing.T) {
	svc := K8sDiagService{}
	if !strings.Contains(svc.InspectPodFailure("checkout-api-789"), "Database connectivity failure") {
		t.Error("esperava diagnostico de conectividade para pod 'api'")
	}
	if !strings.Contains(svc.InspectPodFailure("image-worker-1"), "OOMKilled") {
		t.Error("esperava OOMKilled para pod 'worker'")
	}
	if !strings.Contains(svc.InspectPodFailure("unrelated-pod"), "Readiness Probe is failing") {
		t.Error("esperava fallback para pod desconhecido")
	}
}

func TestSuggestFix_RetornaRemediacaoConhecidaOuFallback(t *testing.T) {
	svc := K8sDiagService{}
	if !strings.Contains(svc.SuggestFix("OOMKilled"), "resources.limits.memory") {
		t.Error("esperava remediacao para OOMKilled")
	}
	if !strings.Contains(svc.SuggestFix("TipoDesconhecido"), "Readiness Probe") {
		t.Error("esperava fallback generico")
	}
}

// --- SecurityScanService ---

func TestRunCheckovScan_ArquivoInexistenteRetornaErro(t *testing.T) {
	svc := SecurityScanService{}
	result := svc.RunCheckovScan("nao-existe.tf")
	if !strings.Contains(result, "not found for scanning") {
		t.Errorf("resultado inesperado: %q", result)
	}
}

func TestValidateOpaPolicies_RegrasDeGovernanca(t *testing.T) {
	svc := SecurityScanService{}
	if !strings.Contains(svc.ValidateOpaPolicies("region = eu-west-1"), "SOBERANIA_DADOS") {
		t.Error("esperava rejeicao por regiao")
	}
	if !strings.Contains(svc.ValidateOpaPolicies("us-east-1 t3.large"), "COST_CONTROL") {
		t.Error("esperava rejeicao por tamanho de instancia")
	}
	if !strings.Contains(svc.ValidateOpaPolicies("us-east-1 cidr 0.0.0.0/0"), "NO_PUBLIC_INGRESS") {
		t.Error("esperava rejeicao por ingress publico")
	}
	if !strings.Contains(svc.ValidateOpaPolicies("us-east-1 t3.micro cidr 10.0.0.0/16"), "OPA PASSED") {
		t.Error("esperava aprovacao")
	}
}

// --- ObservabilityService ---

func TestQueryPrometheusMetrics_DiferenciaPorPalavraChave(t *testing.T) {
	svc := ObservabilityService{}
	if !strings.Contains(svc.QueryPrometheusMetrics("show me the latency"), "latency") {
		t.Error("esperava resultado de latencia")
	}
	if !strings.Contains(svc.QueryPrometheusMetrics("error rate"), "5XX error rate") {
		t.Error("esperava resultado de taxa de erro")
	}
	if !strings.Contains(svc.QueryPrometheusMetrics("cpu usage"), "normal baseline") {
		t.Error("esperava resultado padrao")
	}
}

func TestQueryJaegerTraces_IncluiNomeDoServico(t *testing.T) {
	svc := ObservabilityService{}
	if !strings.Contains(svc.QueryJaegerTraces("checkout-api"), "checkout-api") {
		t.Error("esperava nome do servico no resultado")
	}
}

// --- AiOpsService ---

func TestNlToPromql_TraduzPorPalavraChave(t *testing.T) {
	svc := AiOpsService{}
	if !strings.Contains(svc.NlToPromql("taxa de erro do checkout"), "http_requests_total") {
		t.Error("esperava query de taxa de erro")
	}
	if !strings.Contains(svc.NlToPromql("uso de disco"), "node_filesystem_avail_bytes") {
		t.Error("esperava query de disco")
	}
	if svc.NlToPromql("status geral") != `up{job="kubernetes-pods"}` {
		t.Error("esperava query padrao")
	}
}

func TestPredictiveDiskAlert_DetectaCrescimento(t *testing.T) {
	svc := AiOpsService{}
	if !strings.Contains(svc.PredictiveDiskAlert("crescimento acelerado"), "ALERTA PREDITIVO") {
		t.Error("esperava alerta preditivo")
	}
	if !strings.Contains(svc.PredictiveDiskAlert("uso estavel"), "Padrão de uso normal") {
		t.Error("esperava padrao normal")
	}
}

func TestGenerateGrafanaDashboard_SalvaJsonValido(t *testing.T) {
	cleanupFile(t, "incident_dashboard.json")
	svc := AiOpsService{}

	result := svc.GenerateGrafanaDashboard("Disco cheio em checkout-db")
	if !strings.Contains(result, "Dashboard gerado com sucesso") {
		t.Errorf("resultado inesperado: %q", result)
	}

	content, err := os.ReadFile("incident_dashboard.json")
	if err != nil {
		t.Fatalf("erro lendo dashboard gerado: %v", err)
	}
	if !strings.Contains(string(content), "Disco cheio em checkout-db") || !strings.Contains(string(content), "panels") {
		t.Errorf("conteudo do dashboard inesperado: %s", content)
	}
}

// --- ChatOpsService ---

func TestExecuteTerraform_BloqueiaAcaoDestrutivaSemSenha(t *testing.T) {
	svc := ChatOpsService{}
	if !strings.Contains(svc.ExecuteTerraform("destroy tudo", "None"), "BLOCKED") {
		t.Error("esperava bloqueio sem senha")
	}
	if !strings.Contains(svc.ExecuteTerraform("destroy tudo", "GESTOR-APROVA"), "APPROVED") {
		t.Error("esperava aprovacao com senha correta")
	}
	if !strings.Contains(svc.ExecuteTerraform("terraform plan", "None"), "SUCCESS") {
		t.Error("esperava sucesso para comando nao destrutivo")
	}
}

// --- FileWriterService ---

func TestWriteFile_RemoveCercasMarkdownESalva(t *testing.T) {
	cleanupFile(t, "test-output.tf")
	svc := FileWriterService{}

	result := svc.WriteFile("```hcl\nresource \"aws_s3_bucket\" \"x\" {}\n```", "test-output.tf")
	if !strings.Contains(result, "saved successfully") {
		t.Errorf("resultado inesperado: %q", result)
	}

	content, err := os.ReadFile("test-output.tf")
	if err != nil {
		t.Fatalf("erro lendo arquivo gerado: %v", err)
	}
	if strings.Contains(string(content), "```") {
		t.Error("esperava cercas markdown removidas")
	}
	if !strings.Contains(string(content), "aws_s3_bucket") {
		t.Error("esperava conteudo preservado")
	}
}

// --- RunbookService ---

func TestConsultRunbook_LeRunbookExistente(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(tempDir+"/runbook_db.md", []byte("# Runbook de teste"), 0o644); err != nil {
		t.Fatalf("erro no setup: %v", err)
	}

	svc := RunbookService{DataPath: tempDir}
	if got := svc.ConsultRunbook("db"); got != "# Runbook de teste" {
		t.Errorf("esperava conteudo do runbook, obteve %q", got)
	}
}

func TestConsultRunbook_ServicoSemRunbookRetornaErro(t *testing.T) {
	svc := RunbookService{DataPath: t.TempDir()}
	if got := svc.ConsultRunbook("inexistente"); !strings.Contains(got, "not found") {
		t.Errorf("esperava erro 'not found', obteve %q", got)
	}
}
