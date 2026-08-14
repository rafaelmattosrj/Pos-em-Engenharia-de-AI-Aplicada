package com.trialforge.tiering;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.nio.file.Path;
import java.util.List;
import java.util.concurrent.Callable;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.offset;

/**
 * Cobre os mesmos 4 cenários demonstrados em main() do original
 * (trialforge-model-tiering-prototype.js / .py), usando um {@link FakeOllamaGateway}
 * no lugar do Ollama real — determinístico e sem rede, mas exercitando o fluxo
 * completo (RAG, cascata, orçamento, Approval Gate, auditoria).
 */
class CascadeGatewayTest {

    private static final double[] TEMA0 = {1, 0};
    private static final double[] TEXTO0 = {1, 0};
    private static final double[] TEMA1 = {0, 1};
    private static final double[] TEXTO1 = {0, 1};

    private RagIndex construirIndice(FakeOllamaGateway fake) throws Exception {
        fake.comEmbedding(RagIndex.BANCO_CLAUSULAS.get(0).tema(), TEMA0)
                .comEmbedding(RagIndex.BANCO_CLAUSULAS.get(0).texto(), TEXTO0)
                .comEmbedding(RagIndex.BANCO_CLAUSULAS.get(1).tema(), TEMA1)
                .comEmbedding(RagIndex.BANCO_CLAUSULAS.get(1).texto(), TEXTO1);
        RagIndex ragIndex = new RagIndex(fake, CascadeGateway.MODELO_EMBEDDING);
        ragIndex.prepararIndice();
        return ragIndex;
    }

    // ---------- Cenário 1: Tier 1 resolve sozinho, sem escalar ----------

    @Test
    void processarComCascata_confiancaAlta_resolveNoTier1SemEscalar(@TempDir Path tempDir) throws Exception {
        String pergunta = "pergunta-rotina-tier1";
        FakeOllamaGateway fake = new FakeOllamaGateway()
                .comEmbedding(pergunta, new double[]{1, 0})
                .comEmbedding("resposta-tier1-rotina", new double[]{1, 0})
                .comChat(CascadeGateway.MODELO_TIER1, pergunta, "resposta-tier1-rotina");
        RagIndex ragIndex = construirIndice(fake);
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-A", 0.05);
        AuditTrail auditTrail = new AuditTrail(tempDir.resolve("audit.jsonl"));
        CascadeGateway gateway = new CascadeGateway(fake, ragIndex, orcamento, new FakeApprovalPrompt(true), auditTrail);

        String rascunho = gateway.processarComCascata(pergunta, "estudo-A");

        assertThat(rascunho).isEqualTo("resposta-tier1-rotina");
        assertThat(orcamento.gastoAtual("estudo-A")).isCloseTo(CascadeGateway.CUSTO_TIER1, offset(1e-9));

        var linhas = auditTrail.lerTodas();
        assertThat(linhas).hasSize(1);
        assertThat(linhas.get(0).path("tier_usado").asText()).isEqualTo("Tier 1");
        assertThat(linhas.get(0).path("escalou_cascata").asBoolean()).isFalse();
        assertThat(linhas.get(0).path("status_final").asText()).isEqualTo("aprovado");
    }

    // ---------- Cenário 2: pergunta fora do domínio -> escala pro Tier 2 ----------

    @Test
    void processarComCascata_confiancaBaixaDeBusca_escalaParaTier2(@TempDir Path tempDir) throws Exception {
        String pergunta = "pergunta-fora-dominio-escalada";
        FakeOllamaGateway fake = new FakeOllamaGateway()
                .comEmbedding(pergunta, new double[]{1, 1}) // equidistante dos dois temas, abaixo do limiar
                .comEmbedding("resposta-tier1-fora-dominio", new double[]{1, 1})
                .comEmbedding("resposta-tier2-escalada", new double[]{1, 0})
                .comChat(CascadeGateway.MODELO_TIER1, pergunta, "resposta-tier1-fora-dominio")
                .comChat(CascadeGateway.MODELO_TIER2, pergunta, "resposta-tier2-escalada");
        RagIndex ragIndex = construirIndice(fake);
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-A", 0.05);
        AuditTrail auditTrail = new AuditTrail(tempDir.resolve("audit.jsonl"));
        CascadeGateway gateway = new CascadeGateway(fake, ragIndex, orcamento, new FakeApprovalPrompt(true), auditTrail);

        String rascunho = gateway.processarComCascata(pergunta, "estudo-A");

        assertThat(rascunho).isEqualTo("resposta-tier2-escalada");
        // custo real: Tier1 (tentativa) + Tier2 (escalação) = 0.001 + 0.01
        assertThat(orcamento.gastoAtual("estudo-A"))
                .isCloseTo(CascadeGateway.CUSTO_TIER1 + CascadeGateway.CUSTO_TIER2, offset(1e-9));

        var linhas = auditTrail.lerTodas();
        assertThat(linhas.get(0).path("tier_usado").asText()).isEqualTo("Tier 2 (escalado)");
        assertThat(linhas.get(0).path("escalou_cascata").asBoolean()).isTrue();
        assertThat(linhas.get(0).path("confianca_resposta").isNumber()).isTrue();
    }

    // ---------- Cenário 3: síntese de CSR -> regra fixa pro Tier 2 + Approval Gate ----------

    @Test
    void processarComCascata_sinteseDeCsr_vaiDireitoParaTier2ComAprovacao(@TempDir Path tempDir) throws Exception {
        String pergunta = "Preciso do CSR final agora, por favor.";
        FakeOllamaGateway fake = new FakeOllamaGateway()
                .comEmbedding(pergunta, new double[]{1, 0})
                .comChat(CascadeGateway.MODELO_TIER2, "CSR", "resposta-tier2-csr");
        RagIndex ragIndex = construirIndice(fake);
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-A", 0.05);
        AuditTrail auditTrail = new AuditTrail(tempDir.resolve("audit.jsonl"));
        FakeApprovalPrompt aprovacao = new FakeApprovalPrompt(true);
        CascadeGateway gateway = new CascadeGateway(fake, ragIndex, orcamento, aprovacao, auditTrail);

        String rascunho = gateway.processarComCascata(pergunta, "estudo-A");

        assertThat(rascunho).isEqualTo("resposta-tier2-csr");
        assertThat(aprovacao.chamadas).isEqualTo(1);
        assertThat(aprovacao.ultimoRascunho).isEqualTo("resposta-tier2-csr");
        // regra fixa: só Tier 2 é chamado, sem cascata (nunca tenta Tier 1 primeiro)
        assertThat(orcamento.gastoAtual("estudo-A")).isCloseTo(CascadeGateway.CUSTO_TIER2, offset(1e-9));

        var linhas = auditTrail.lerTodas();
        assertThat(linhas.get(0).path("tier_usado").asText()).isEqualTo("Tier 2");
        assertThat(linhas.get(0).path("escalou_cascata").asBoolean()).isFalse();
        assertThat(linhas.get(0).path("aprovado").asBoolean()).isTrue();
        assertThat(linhas.get(0).path("confianca_resposta").isNull()).isTrue();
    }

    @Test
    void processarComCascata_sinteseDeCsrReprovada_marcaRejeitado(@TempDir Path tempDir) throws Exception {
        String pergunta = "Preciso do CSR final agora, por favor.";
        FakeOllamaGateway fake = new FakeOllamaGateway()
                .comEmbedding(pergunta, new double[]{1, 0})
                .comChat(CascadeGateway.MODELO_TIER2, "CSR", "resposta-tier2-csr");
        RagIndex ragIndex = construirIndice(fake);
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-A", 0.05);
        AuditTrail auditTrail = new AuditTrail(tempDir.resolve("audit.jsonl"));
        CascadeGateway gateway = new CascadeGateway(fake, ragIndex, orcamento, new FakeApprovalPrompt(false), auditTrail);

        gateway.processarComCascata(pergunta, "estudo-A");

        var linhas = auditTrail.lerTodas();
        assertThat(linhas.get(0).path("aprovado").asBoolean()).isFalse();
        assertThat(linhas.get(0).path("status_final").asText()).isEqualTo("rejeitado");
    }

    // ---------- Cenário 4: estudo com orçamento baixo -> bloqueia antes de chamar o modelo ----------

    @Test
    void processarComCascata_orcamentoInsuficiente_bloqueiaAntesDeChamarModelo(@TempDir Path tempDir) throws Exception {
        String pergunta = "pergunta-orcamento-baixo";
        FakeOllamaGateway fake = new FakeOllamaGateway(); // nenhum chat/embed configurado: se for chamado, o teste falha
        RagIndex ragIndex = construirIndice(fake);
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-B", 0.005); // menor que CUSTO_TIER1+CUSTO_TIER2 = 0.011
        AuditTrail auditTrail = new AuditTrail(tempDir.resolve("audit.jsonl"));
        CascadeGateway gateway = new CascadeGateway(fake, ragIndex, orcamento, new FakeApprovalPrompt(true), auditTrail);

        String rascunho = gateway.processarComCascata(pergunta, "estudo-B");

        assertThat(rascunho).isNull();
        assertThat(orcamento.gastoAtual("estudo-B")).isEqualTo(0.0);

        var linhas = auditTrail.lerTodas();
        assertThat(linhas).hasSize(1);
        assertThat(linhas.get(0).path("status_final").asText()).isEqualTo("bloqueado_por_orcamento");
    }

    // ---------- Os 4 cenários em sequência, como em main(): a trilha de auditoria confirma tudo ----------

    @Test
    void quatroCenariosEmSequencia_trilhaDeAuditoriaConfirmaOsTresComportamentos(@TempDir Path tempDir) throws Exception {
        String perguntaRotina = "pergunta-rotina-tier1";
        String perguntaForaDominio = "pergunta-fora-dominio-escalada";
        String perguntaCsr = "Preciso do CSR final agora, por favor.";
        String perguntaBloqueada = "pergunta-orcamento-baixo";

        FakeOllamaGateway fake = new FakeOllamaGateway()
                .comEmbedding(perguntaRotina, new double[]{1, 0})
                .comEmbedding("resposta-tier1-rotina", new double[]{1, 0})
                .comChat(CascadeGateway.MODELO_TIER1, perguntaRotina, "resposta-tier1-rotina")
                .comEmbedding(perguntaForaDominio, new double[]{1, 1})
                .comEmbedding("resposta-tier1-fora-dominio", new double[]{1, 1})
                .comEmbedding("resposta-tier2-escalada", new double[]{1, 0})
                .comChat(CascadeGateway.MODELO_TIER1, perguntaForaDominio, "resposta-tier1-fora-dominio")
                .comChat(CascadeGateway.MODELO_TIER2, perguntaForaDominio, "resposta-tier2-escalada")
                .comEmbedding(perguntaCsr, new double[]{1, 0})
                .comChat(CascadeGateway.MODELO_TIER2, "CSR", "resposta-tier2-csr");

        RagIndex ragIndex = construirIndice(fake);
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-A", 0.05);
        orcamento.definirOrcamento("estudo-B", 0.005);
        AuditTrail auditTrail = new AuditTrail(tempDir.resolve("audit.jsonl"));
        CascadeGateway gateway = new CascadeGateway(fake, ragIndex, orcamento, new FakeApprovalPrompt(true), auditTrail);

        gateway.processarComCascata(perguntaRotina, "estudo-A");
        gateway.processarComCascata(perguntaForaDominio, "estudo-A");
        gateway.processarComCascata(perguntaCsr, "estudo-A");
        gateway.processarComCascata(perguntaBloqueada, "estudo-B");

        AuditTrailVerifier.verificarEImprimir(auditTrail); // não deve lançar
    }

    // ---------- Volume concorrente: réplica de simularVolumeConcorrente, com fake gateway ----------

    @Test
    void volumeConcorrente_reservarOrcamentoSeguraOOrcamentoPorEstudo(@TempDir Path tempDir) throws Exception {
        String perguntaRotina = "pergunta-rotina-tier1";
        FakeOllamaGateway fake = new FakeOllamaGateway()
                .comEmbeddingPadrao(new double[]{1, 0})
                .comChatPadrao("resposta-padrao");
        RagIndex ragIndex = construirIndice(fake);

        OrcamentoManager orcamento = new OrcamentoManager();
        double custoPiorCaso = CascadeGateway.CUSTO_TIER1 + CascadeGateway.CUSTO_TIER2;
        double limiteApertado = 2 * custoPiorCaso + 0.0005; // cabe exatamente 2 reservas
        orcamento.definirOrcamento("estudo-F", limiteApertado);

        AuditTrail auditTrail = new AuditTrail(tempDir.resolve("audit.jsonl"));
        CascadeGateway gateway = new CascadeGateway(fake, ragIndex, orcamento, new FakeApprovalPrompt(true), auditTrail);

        List<Callable<String>> requisicoes = new java.util.ArrayList<>();
        for (int i = 0; i < 5; i++) {
            requisicoes.add(() -> gateway.processarComCascata(perguntaRotina, "estudo-F"));
        }

        List<VolumeSimulator.ResultadoEstudo> resultados =
                VolumeSimulator.simular(gateway, orcamento, List.of("estudo-F"), requisicoes, 5);

        assertThat(resultados).hasSize(1);
        assertThat(resultados.get(0).estourou()).isFalse();
        assertThat(orcamento.gastoAtual("estudo-F")).isLessThanOrEqualTo(limiteApertado + 1e-9);
    }
}
