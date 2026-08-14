// Package handler expõe as ferramentas do pacote service como endpoints
// HTTP — equivalente aos controllers Spring do porte Java irmão
// (nexus-tools-java) e, em última instância, às funções decoradas com
// @tool do CrewAI na versão Python original.
package handler

import (
	"encoding/json"
	"net/http"

	"nexus-tools/service"
)

type toolResponse struct {
	Result string `json:"result"`
}

func writeResult(w http.ResponseWriter, result string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toolResponse{Result: result})
}

func decodeOrBadRequest[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "corpo da requisicao invalido"})
		var zero T
		return zero, false
	}
	return body, true
}

// Handlers agrupa todos os handlers HTTP das ferramentas Nexus.
type Handlers struct {
	K8sOps        service.K8sOpsService
	K8sDiag       service.K8sDiagService
	SecurityScan  service.SecurityScanService
	Observability service.ObservabilityService
	AiOps         service.AiOpsService
	ChatOps       service.ChatOpsService
	Governance    service.GovernanceService
	PolicyRag     service.PolicyRagService
	FileWriter    service.FileWriterService
	Runbook       service.RunbookService
}

// Register registra todas as rotas /tools/* no mux informado.
func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /tools/generate-k8s-manifest", h.generateK8sManifest)
	mux.HandleFunc("POST /tools/apply-k8s-manifest", h.applyK8sManifest)
	mux.HandleFunc("POST /tools/analyze-canary-metrics", h.analyzeCanaryMetrics)
	mux.HandleFunc("POST /tools/inspect-pod-failure", h.inspectPodFailure)
	mux.HandleFunc("POST /tools/suggest-fix", h.suggestFix)
	mux.HandleFunc("POST /tools/run-checkov-scan", h.runCheckovScan)
	mux.HandleFunc("POST /tools/validate-opa-policies", h.validateOpaPolicies)
	mux.HandleFunc("POST /tools/query-prometheus-metrics", h.queryPrometheusMetrics)
	mux.HandleFunc("POST /tools/query-jaeger-traces", h.queryJaegerTraces)
	mux.HandleFunc("POST /tools/nl-to-promql", h.nlToPromql)
	mux.HandleFunc("POST /tools/predictive-disk-alert", h.predictiveDiskAlert)
	mux.HandleFunc("POST /tools/generate-grafana-dashboard", h.generateGrafanaDashboard)
	mux.HandleFunc("POST /tools/execute-terraform", h.executeTerraform)
	mux.HandleFunc("POST /tools/triage-security-vulnerabilities", h.triageSecurityVulnerabilities)
	mux.HandleFunc("POST /tools/optimize-cicd-pipeline", h.optimizeCicdPipeline)
	mux.HandleFunc("POST /tools/analyze-finops-costs", h.analyzeFinopsCosts)
	mux.HandleFunc("POST /tools/check-compliance-rules", h.checkComplianceRules)
	mux.HandleFunc("POST /tools/write-file", h.writeFile)
	mux.HandleFunc("POST /tools/consult-runbook", h.consultRunbook)
}

func (h *Handlers) generateK8sManifest(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		AppName  string `json:"appName"`
		Replicas int    `json:"replicas"`
		Port     int    `json:"port"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.K8sOps.GenerateK8sManifest(body.AppName, body.Replicas, body.Port))
}

func (h *Handlers) applyK8sManifest(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		Filename string `json:"filename"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.K8sOps.ApplyK8sManifest(body.Filename))
}

func (h *Handlers) analyzeCanaryMetrics(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		MetricsData string `json:"metricsData"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.K8sOps.AnalyzeCanaryMetrics(body.MetricsData))
}

func (h *Handlers) inspectPodFailure(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		PodName string `json:"podName"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.K8sDiag.InspectPodFailure(body.PodName))
}

func (h *Handlers) suggestFix(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		IssueType string `json:"issueType"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.K8sDiag.SuggestFix(body.IssueType))
}

func (h *Handlers) runCheckovScan(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		Filename string `json:"filename"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.SecurityScan.RunCheckovScan(body.Filename))
}

func (h *Handlers) validateOpaPolicies(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		Content string `json:"content"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.SecurityScan.ValidateOpaPolicies(body.Content))
}

func (h *Handlers) queryPrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		Query string `json:"query"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.Observability.QueryPrometheusMetrics(body.Query))
}

func (h *Handlers) queryJaegerTraces(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		ServiceName string `json:"serviceName"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.Observability.QueryJaegerTraces(body.ServiceName))
}

func (h *Handlers) nlToPromql(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		NaturalLanguageQuery string `json:"naturalLanguageQuery"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.AiOps.NlToPromql(body.NaturalLanguageQuery))
}

func (h *Handlers) predictiveDiskAlert(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		MetricsHistory string `json:"metricsHistory"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.AiOps.PredictiveDiskAlert(body.MetricsHistory))
}

func (h *Handlers) generateGrafanaDashboard(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		IncidentContext string `json:"incidentContext"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.AiOps.GenerateGrafanaDashboard(body.IncidentContext))
}

func (h *Handlers) executeTerraform(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		Command         string `json:"command"`
		ManagerPassword string `json:"managerPassword"`
	}](w, r)
	if !ok {
		return
	}
	password := body.ManagerPassword
	if password == "" {
		password = "None"
	}
	writeResult(w, h.ChatOps.ExecuteTerraform(body.Command, password))
}

func (h *Handlers) triageSecurityVulnerabilities(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		TrivyJSONReport string `json:"trivyJsonReport"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.Governance.TriageSecurityVulnerabilities(body.TrivyJSONReport))
}

func (h *Handlers) optimizeCicdPipeline(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		WorkflowYamlContent string `json:"workflowYamlContent"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.Governance.OptimizeCicdPipeline(body.WorkflowYamlContent))
}

func (h *Handlers) analyzeFinopsCosts(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		CurrentResourcesInventory string `json:"currentResourcesInventory"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.Governance.AnalyzeFinopsCosts(body.CurrentResourcesInventory))
}

func (h *Handlers) checkComplianceRules(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		Query string `json:"query"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.PolicyRag.CheckComplianceRules(body.Query))
}

func (h *Handlers) writeFile(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		Content  string `json:"content"`
		Filename string `json:"filename"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.FileWriter.WriteFile(body.Content, body.Filename))
}

func (h *Handlers) consultRunbook(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeOrBadRequest[struct {
		ServiceName string `json:"serviceName"`
	}](w, r)
	if !ok {
		return
	}
	writeResult(w, h.Runbook.ConsultRunbook(body.ServiceName))
}
