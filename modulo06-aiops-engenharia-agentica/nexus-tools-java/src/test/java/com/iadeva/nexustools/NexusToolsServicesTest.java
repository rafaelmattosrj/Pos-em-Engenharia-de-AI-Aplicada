package com.iadeva.nexustools;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.iadeva.nexustools.service.*;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Testes unitários das ferramentas — equivalente à lógica testável de
 * tools/*.py (as tools decoradas com @tool do CrewAI não tinham testes
 * automatizados na versão Python original; esta suíte cobre os principais
 * ramos de decisão de cada uma).
 */
class NexusToolsServicesTest {

    private final java.util.List<String> filesToCleanup = new java.util.ArrayList<>();

    @AfterEach
    void cleanup() {
        for (String f : filesToCleanup) {
            new File(f).delete();
        }
        filesToCleanup.clear();
    }

    // --- K8sOpsService ---

    @Test
    void generateK8sManifest_escreveArquivoComConteudoEsperado() throws IOException {
        K8sOpsService service = new K8sOpsService();
        filesToCleanup.add("checkout-k8s.yaml");

        String result = service.generateK8sManifest("checkout", 3, 8080);

        assertThat(result).contains("successfully generated");
        String content = Files.readString(Path.of("checkout-k8s.yaml"));
        assertThat(content).contains("name: checkout").contains("replicas: 3").contains("containerPort: 8080");
    }

    @Test
    void applyK8sManifest_arquivoInexistenteRetornaErro() {
        K8sOpsService service = new K8sOpsService();
        String result = service.applyK8sManifest("arquivo-que-nao-existe.yaml");
        assertThat(result).contains("was not found to apply");
    }

    @Test
    void analyzeCanaryMetrics_erroDisparaRollback() {
        K8sOpsService service = new K8sOpsService();
        assertThat(service.analyzeCanaryMetrics("error_rate > 5%")).contains("ROLLBACK");
        assertThat(service.analyzeCanaryMetrics("all metrics stable")).contains("PROCEED");
    }

    // --- K8sDiagService ---

    @Test
    void inspectPodFailure_diferenciaPorNomeDoPod() {
        K8sDiagService service = new K8sDiagService();
        assertThat(service.inspectPodFailure("checkout-api-789")).contains("Database connectivity failure");
        assertThat(service.inspectPodFailure("image-worker-1")).contains("OOMKilled");
        assertThat(service.inspectPodFailure("unrelated-pod")).contains("Readiness Probe is failing");
    }

    @Test
    void suggestFix_retornaRemediacaoConhecidaOuFallback() {
        K8sDiagService service = new K8sDiagService();
        assertThat(service.suggestFix("OOMKilled")).contains("resources.limits.memory");
        assertThat(service.suggestFix("TipoDesconhecido")).contains("Readiness Probe");
    }

    // --- SecurityScanService ---

    @Test
    void runCheckovScan_arquivoInexistenteRetornaErro() {
        SecurityScanService service = new SecurityScanService();
        assertThat(service.runCheckovScan("nao-existe.tf")).contains("not found for scanning");
    }

    @Test
    void validateOpaPolicies_regrasDeGovernanca() {
        SecurityScanService service = new SecurityScanService();
        assertThat(service.validateOpaPolicies("region = eu-west-1")).contains("SOBERANIA_DADOS");
        assertThat(service.validateOpaPolicies("us-east-1 t3.large")).contains("COST_CONTROL");
        assertThat(service.validateOpaPolicies("us-east-1 cidr 0.0.0.0/0")).contains("NO_PUBLIC_INGRESS");
        assertThat(service.validateOpaPolicies("us-east-1 t3.micro cidr 10.0.0.0/16")).contains("OPA PASSED");
    }

    // --- ObservabilityService ---

    @Test
    void queryPrometheusMetrics_diferenciaPorPalavraChave() {
        ObservabilityService service = new ObservabilityService();
        assertThat(service.queryPrometheusMetrics("show me the latency")).contains("latency");
        assertThat(service.queryPrometheusMetrics("error rate")).contains("5XX error rate");
        assertThat(service.queryPrometheusMetrics("cpu usage")).contains("normal baseline");
    }

    @Test
    void queryJaegerTraces_incluiNomeDoServico() {
        ObservabilityService service = new ObservabilityService();
        assertThat(service.queryJaegerTraces("checkout-api")).contains("checkout-api");
    }

    // --- AiOpsService ---

    @Test
    void nlToPromql_traduzPorPalavraChave() {
        AiOpsService service = new AiOpsService(new ObjectMapper());
        assertThat(service.nlToPromql("taxa de erro do checkout")).contains("http_requests_total");
        assertThat(service.nlToPromql("uso de disco")).contains("node_filesystem_avail_bytes");
        assertThat(service.nlToPromql("status geral")).isEqualTo("up{job=\"kubernetes-pods\"}");
    }

    @Test
    void predictiveDiskAlert_detectaCrescimento() {
        AiOpsService service = new AiOpsService(new ObjectMapper());
        assertThat(service.predictiveDiskAlert("crescimento acelerado")).contains("ALERTA PREDITIVO");
        assertThat(service.predictiveDiskAlert("uso estavel")).contains("Padrão de uso normal");
    }

    @Test
    void generateGrafanaDashboard_salvaJsonValido() throws IOException {
        AiOpsService service = new AiOpsService(new ObjectMapper());
        filesToCleanup.add("incident_dashboard.json");

        String result = service.generateGrafanaDashboard("Disco cheio em checkout-db");

        assertThat(result).contains("Dashboard gerado com sucesso");
        String json = Files.readString(Path.of("incident_dashboard.json"));
        assertThat(json).contains("Disco cheio em checkout-db").contains("panels");
    }

    // --- ChatOpsService ---

    @Test
    void executeTerraform_bloqueiaAcaoDestrutivaSemSenha() {
        ChatOpsService service = new ChatOpsService();
        assertThat(service.executeTerraform("destroy tudo", "None")).contains("BLOCKED");
        assertThat(service.executeTerraform("destroy tudo", "GESTOR-APROVA")).contains("APPROVED");
        assertThat(service.executeTerraform("terraform plan", "None")).contains("SUCCESS");
    }

    // --- FileWriterService ---

    @Test
    void writeFile_removeCercasMarkdownESalva() throws IOException {
        FileWriterService service = new FileWriterService();
        filesToCleanup.add("test-output.tf");

        String result = service.writeFile("```hcl\nresource \"aws_s3_bucket\" \"x\" {}\n```", "test-output.tf");

        assertThat(result).contains("saved successfully");
        String content = Files.readString(Path.of("test-output.tf"));
        assertThat(content).doesNotContain("```").contains("aws_s3_bucket");
    }

    // --- RunbookService ---

    @Test
    void consultRunbook_leRunbookExistente(@TempDir Path tempDir) throws IOException {
        Files.writeString(tempDir.resolve("runbook_db.md"), "# Runbook de teste");

        RunbookService service = new RunbookService(tempDir.toString());
        assertThat(service.consultRunbook("db")).isEqualTo("# Runbook de teste");
    }

    @Test
    void consultRunbook_servicoSemRunbookRetornaErro(@TempDir Path tempDir) {
        RunbookService service = new RunbookService(tempDir.toString());
        assertThat(service.consultRunbook("inexistente")).contains("not found");
    }
}
