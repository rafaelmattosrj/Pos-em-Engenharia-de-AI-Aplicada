package com.trialforge.gateway;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Cobre os 5 caminhos do roteiro de demo (Main.java / trialforge-gateway-
 * prototype.js) com embeddings/chat/approval controlados — sem depender de
 * um Ollama real nem de stdin.
 */
class ProcessorTest {

    private static final String PERGUNTA_ROTINA = "Quais são as regras de assentimento pra menores nesse estudo?";
    private static final String PERGUNTA_PARAFRASE = "O assentimento dos menores de idade é obrigatório nesse estudo?";
    private static final String PERGUNTA_CSR = "Preciso da síntese do CSR final desse estudo.";
    private static final String PERGUNTA_TEMA_DIFERENTE = "Qual o prazo de armazenamento das amostras biológicas coletadas nesse estudo?";
    private static final String PERGUNTA_PROTOCOLO = "Qual é o critério de idade mínima pra participar desse estudo?";

    /**
     * Monta um Processor com embeddings controlados que reproduzem os 5
     * caminhos do roteiro de demo sem precisar de um Ollama real: cada
     * cláusula real dos 3 índices (Indices.INDICES) recebe um vetor de teste
     * ortogonal por domínio (icf~{1,0,0}, protocolo~{0,1,0}, csr~{0,0,1}), e
     * cada pergunta de teste recebe o vetor que produz o comportamento
     * esperado daquele cenário.
     */
    private Processor montarProcessor(Path tempDir, FakeEmbedder embedder, FakeChatStreamer chat, FakeApprover approver,
            ByteArrayOutputStream saidaBuffer) throws Exception {
        return montarProcessor(tempDir, embedder, chat, approver, saidaBuffer, new SemanticCache());
    }

    private Processor montarProcessor(Path tempDir, FakeEmbedder embedder, FakeChatStreamer chat, FakeApprover approver,
            ByteArrayOutputStream saidaBuffer, SemanticCache cache) throws Exception {
        embedder
                .com(Indices.INDICES.get("icf").get(0).tema(), List.of(1.0, 0.0, 0.0))
                .com(Indices.INDICES.get("icf").get(0).texto(), List.of(1.0, 0.0, 0.0))
                .com(Indices.INDICES.get("icf").get(1).tema(), List.of(0.9, 0.1, 0.0))
                .com(Indices.INDICES.get("icf").get(1).texto(), List.of(0.9, 0.1, 0.0))
                .com(Indices.INDICES.get("protocolo").get(0).tema(), List.of(0.0, 1.0, 0.0))
                .com(Indices.INDICES.get("protocolo").get(0).texto(), List.of(0.0, 1.0, 0.0))
                .com(Indices.INDICES.get("protocolo").get(1).tema(), List.of(0.0, 0.9, 0.1))
                .com(Indices.INDICES.get("protocolo").get(1).texto(), List.of(0.0, 0.9, 0.1))
                .com(Indices.INDICES.get("csr").get(0).tema(), List.of(0.0, 0.0, 1.0))
                .com(Indices.INDICES.get("csr").get(0).texto(), List.of(0.0, 0.0, 1.0))
                .com(Indices.INDICES.get("csr").get(1).tema(), List.of(0.0, 0.1, 0.9))
                .com(Indices.INDICES.get("csr").get(1).texto(), List.of(0.0, 0.1, 0.9));

        Map<String, IndicePreparado> preparados = RagSearch.prepararIndices(embedder, null);

        PrintStream out = new PrintStream(saidaBuffer, true, StandardCharsets.UTF_8);
        return new Processor(
                embedder, chat, preparados,
                cache,
                new AuditTrail(tempDir.resolve("audit-trail.jsonl").toString()),
                approver,
                out,
                "gemma4:e2b", "gemma4:latest",
                0.75, 0.7);
    }

    // #1 do roteiro: confiança alta (embedding igual ao tema da cláusula do
    // índice icf) — sem cache (cache vazio), sem Approval Gate, modelo barato.
    @Test
    void rotinaComAltaConfiancaNaoAcionaGateNemCache(@TempDir Path tempDir) throws Exception {
        FakeEmbedder embedder = new FakeEmbedder().com(PERGUNTA_ROTINA, List.of(1.0, 0.0, 0.0));
        FakeChatStreamer chat = new FakeChatStreamer("resposta de rotina");
        FakeApprover approver = new FakeApprover();
        Processor p = montarProcessor(tempDir, embedder, chat, approver, new ByteArrayOutputStream());

        String resposta = p.processarRequisicao(PERGUNTA_ROTINA);

        assertThat(resposta).isEqualTo("resposta de rotina");
        assertThat(chat.chamadas).isEqualTo(1);
        assertThat(chat.ultimoModelo).isEqualTo("gemma4:e2b");
        assertThat(approver.chamadas).isZero();
    }

    // #2 do roteiro: paráfrase da #1 — mesmo embedding da pergunta anterior
    // já cacheada, deve bater no Semantic Cache e NUNCA chamar o modelo.
    @Test
    void parafraseUsaSemanticCache(@TempDir Path tempDir) throws Exception {
        FakeEmbedder embedder = new FakeEmbedder()
                .com(PERGUNTA_ROTINA, List.of(1.0, 0.0, 0.0))
                .com(PERGUNTA_PARAFRASE, List.of(1.0, 0.0, 0.0));
        FakeChatStreamer chat = new FakeChatStreamer("resposta de rotina");
        FakeApprover approver = new FakeApprover();
        Processor p = montarProcessor(tempDir, embedder, chat, approver, new ByteArrayOutputStream());

        p.processarRequisicao(PERGUNTA_ROTINA);
        int chamadasAntes = chat.chamadas;

        String resposta = p.processarRequisicao(PERGUNTA_PARAFRASE);

        assertThat(resposta).isEqualTo("resposta de rotina");
        assertThat(chat.chamadas).isEqualTo(chamadasAntes);
    }

    // #3 do roteiro: síntese de CSR — sempre modelo caro, sempre Approval
    // Gate, mesmo com confiança alta.
    @Test
    void sinteseCsrSempreAcionaGateEModeloCaro(@TempDir Path tempDir) throws Exception {
        FakeEmbedder embedder = new FakeEmbedder().com(PERGUNTA_CSR, List.of(0.0, 0.0, 1.0));
        FakeChatStreamer chat = new FakeChatStreamer("resposta de csr");
        FakeApprover approver = new FakeApprover(true);
        Processor p = montarProcessor(tempDir, embedder, chat, approver, new ByteArrayOutputStream());

        String resposta = p.processarRequisicao(PERGUNTA_CSR);

        assertThat(resposta).isEqualTo("resposta de csr");
        assertThat(chat.ultimoModelo).isEqualTo("gemma4:latest");
        assertThat(approver.chamadas).isEqualTo(1);
    }

    // #4 do roteiro: pergunta de tema bem diferente de qualquer índice —
    // Agentic RAG esgota as 3 iterações, Confidence Threshold escala pro
    // Approval Gate por confiança baixa.
    @Test
    void temaDiferenteEscalaGatePorConfiancaBaixa(@TempDir Path tempDir) throws Exception {
        FakeEmbedder embedder = new FakeEmbedder().com(PERGUNTA_TEMA_DIFERENTE, List.of(0.0, 0.0, -1.0));
        FakeChatStreamer chat = new FakeChatStreamer("resposta incerta");
        FakeApprover approver = new FakeApprover(false);
        Processor p = montarProcessor(tempDir, embedder, chat, approver, new ByteArrayOutputStream());

        String resposta = p.processarRequisicao(PERGUNTA_TEMA_DIFERENTE);

        assertThat(resposta).isNull();
        assertThat(approver.chamadas).isEqualTo(1);
    }

    // #5 do roteiro: pergunta de critério de protocolo — Multi-Index roteia
    // pro índice "protocolo" e converge já na 1ª iteração, sem gate.
    @Test
    void protocoloConvergeNaPrimeiraIteracaoSemGate(@TempDir Path tempDir) throws Exception {
        FakeEmbedder embedder = new FakeEmbedder().com(PERGUNTA_PROTOCOLO, List.of(0.0, 1.0, 0.0));
        FakeChatStreamer chat = new FakeChatStreamer("resposta de protocolo");
        FakeApprover approver = new FakeApprover();
        Processor p = montarProcessor(tempDir, embedder, chat, approver, new ByteArrayOutputStream());

        String resposta = p.processarRequisicao(PERGUNTA_PROTOCOLO);

        assertThat(resposta).isEqualTo("resposta de protocolo");
        assertThat(approver.chamadas).isZero();
    }

    // Rejeição no Approval Gate: a resposta não deve ser adicionada ao
    // Semantic Cache (senão uma rejeição vazaria pra respostas futuras).
    @Test
    void rejeicaoNoGateNaoAlimentaCache(@TempDir Path tempDir) throws Exception {
        FakeEmbedder embedder = new FakeEmbedder().com(PERGUNTA_TEMA_DIFERENTE, List.of(0.0, 0.0, -1.0));
        FakeChatStreamer chat = new FakeChatStreamer("rascunho rejeitado");
        FakeApprover approver = new FakeApprover(false);
        SemanticCache cache = new SemanticCache();
        Processor p = montarProcessor(tempDir, embedder, chat, approver, new ByteArrayOutputStream(), cache);

        p.processarRequisicao(PERGUNTA_TEMA_DIFERENTE);

        assertThat(cache.tamanho()).isZero();
    }
}
