package com.trialforge.gateway;

import com.trialforge.gateway.ollama.OllamaClient;

import java.io.File;
import java.util.List;
import java.util.Map;

/**
 * Demonstracao do Gateway do TrialForge — porte Java de
 * trialforge-gateway-prototype.js / trialforge_gateway_prototype.py (Modulo
 * 8, Modulo 4.5 do curso de Arquitetura de Sistemas com IA). Ver README.md
 * para o que foi mantido 1:1 e o que foi adaptado.
 */
public class Main {

    // Os dois limiares abaixo foram calibrados testando de verdade contra o
    // nomic-embed-text em perguntas curtas em portugues (Modulo 4.3: "nao
    // existe limiar universal, se calibra por tipo de pergunta"). Nesse par
    // modelo+idioma, parafrases proximas reaproveitando termos do dominio
    // ficaram em ~0.82 de similaridade, enquanto perguntas de tema totalmente
    // diferente ficaram em ~0.64-0.65 — o corte abaixo fica no meio desse
    // intervalo, nao e um valor padrao de mercado.
    private static final double LIMIAR_CACHE = 0.75;
    private static final double LIMIAR_CONFIANCA = 0.7;

    public static void main(String[] args) {
        String trilhaAuditoria = new File("audit-trail.jsonl").getAbsolutePath();
        try {
            run(trilhaAuditoria);
        } catch (Exception erro) {
            System.out.println("[Erro não tratado] " + erro.getMessage());
            try {
                new AuditTrail(trilhaAuditoria).registrar(Map.of(
                        "erro", String.valueOf(erro.getMessage()),
                        "status_final", "falha_tecnica"));
            } catch (Exception ignorado) {
                // trilha indisponivel nao deve mascarar o erro original
            }
            System.exit(1);
        }
    }

    private static void run(String trilhaAuditoria) throws Exception {
        if (!autoTeste()) {
            throw new IllegalStateException("Testes puros falharam — corrija antes de chamar o modelo.");
        }

        String baseUrl = System.getenv().getOrDefault("OLLAMA_BASE_URL", "http://localhost:11434");
        String modeloBarato = System.getenv().getOrDefault("OLLAMA_MODELO_BARATO", "gemma4:e2b");
        String modeloCaro = System.getenv().getOrDefault("OLLAMA_MODELO_CARO", "gemma4:latest");
        String modeloEmbedding = System.getenv().getOrDefault("OLLAMA_MODELO_EMBEDDING", "nomic-embed-text");

        OllamaClient client = new OllamaClient(baseUrl, modeloEmbedding);
        client.setLogger(msg -> System.out.println(msg));

        Map<String, IndicePreparado> preparados = RagSearch.prepararIndices(client, msg -> System.out.println(msg));

        Processor processor = new Processor(
                client, client, preparados,
                new SemanticCache(),
                new AuditTrail(trilhaAuditoria),
                new ApprovalGate(System.in, System.out),
                System.out,
                modeloBarato, modeloCaro,
                LIMIAR_CACHE, LIMIAR_CONFIANCA);

        List<String> perguntas = List.of(
                // 1) Pergunta de rotina — gera de verdade, depois popula o Semantic Cache
                "Quais são as regras de assentimento pra menores nesse estudo?",
                // 2) Paráfrase da pergunta 1, reaproveitando os termos do domínio —
                // deve bater no Semantic Cache
                "O assentimento dos menores de idade é obrigatório nesse estudo?",
                // 3) Síntese de CSR — sempre modelo caro + sempre Approval Gate, nunca cache
                "Preciso da síntese do CSR final desse estudo.",
                // 4) Pergunta de tema bem diferente de qualquer índice — nenhuma das 3
                // iterações do Agentic RAG encontra confiança suficiente (nem ampliando
                // o índice, nem cruzando os 3 domínios), então esgota o limite e o
                // Confidence Threshold escala pro Approval Gate
                "Qual o prazo de armazenamento das amostras biológicas coletadas nesse estudo?",
                // 5) Pergunta sobre critério de protocolo — Multi-Index roteia pro
                // índice "protocolo" (não "icf"), e Hybrid Search acha a cláusula com
                // confiança já na 1ª iteração
                "Qual é o critério de idade mínima pra participar desse estudo?");

        for (String pergunta : perguntas) {
            processor.processarRequisicao(pergunta);
        }

        new AuditTrail(trilhaAuditoria).verificarTrilhaAuditoria(modeloCaro, LIMIAR_CONFIANCA, msg -> System.out.println(msg));
    }

    /**
     * Roda os mesmos testes puros (sem rede) que os dois originais executam
     * no inicio de main() — determinismo do classificador de intencao, da
     * matematica do cosseno, do BM25 e da fusao RRF — antes de qualquer
     * chamada ao Ollama. Os mesmos cenarios tambem viram testes JUnit em
     * src/test; aqui a funcao reproduz a MESMA saida de console e o MESMO
     * comportamento de abortar antes de tocar a rede.
     */
    private static boolean autoTeste() {
        System.out.println("== Testes: classificarIntencao + similaridadeCosseno (puros, sem rede) ==");
        int passou = 0;
        int total = 0;

        Object[][] casosIntencao = {
                {"Preciso da síntese do CSR final desse estudo.", "sintese_csr"},
                {"Quero o relatório final do estudo.", "sintese_csr"},
                {"Como os eventos adversos aparecem no relatório final?", "sintese_csr"},
                {"Qual é o critério de idade mínima pra participar desse estudo?", "consulta_protocolo"},
                {"Quais são os critérios de exclusão desse protocolo?", "consulta_protocolo"},
                {"Quais são as regras de assentimento pra menores?", "consulta_icf"},
                {"Qual o prazo de armazenamento das amostras biológicas?", "consulta_icf"},
        };
        for (Object[] caso : casosIntencao) {
            total++;
            String pergunta = (String) caso[0];
            String esperado = (String) caso[1];
            String resultado = IntentClassifier.classificarIntencao(pergunta);
            boolean ok = resultado.equals(esperado);
            System.out.printf("  [%s] classificarIntencao(\"%s\") -> %s (esperado %s)%n",
                    ok ? "OK" : "FALHOU", pergunta, resultado, esperado);
            if (ok) {
                passou++;
            }
        }

        List<Double> vetorA = List.of(1.0, 0.0, 0.0);
        List<Double> vetorB = List.of(1.0, 0.0, 0.0);
        List<Double> vetorOrtogonal = List.of(0.0, 1.0, 0.0);
        List<Double> vetorOposto = List.of(-1.0, 0.0, 0.0);
        Object[][] casosCosseno = {
                {"vetores idênticos", vetorA, vetorB, 1.0},
                {"vetores ortogonais", vetorA, vetorOrtogonal, 0.0},
                {"vetores opostos", vetorA, vetorOposto, -1.0},
        };
        for (Object[] caso : casosCosseno) {
            total++;
            String nome = (String) caso[0];
            @SuppressWarnings("unchecked")
            List<Double> a = (List<Double>) caso[1];
            @SuppressWarnings("unchecked")
            List<Double> b = (List<Double>) caso[2];
            double esperado = (double) caso[3];
            double resultado = CosineSimilarity.similaridadeCosseno(a, b);
            boolean ok = Math.abs(resultado - esperado) < 1e-9;
            System.out.printf("  [%s] similaridadeCosseno(%s) -> %.3f (esperado %s)%n",
                    ok ? "OK" : "FALHOU", nome, resultado, esperado);
            if (ok) {
                passou++;
            }
        }

        System.out.println("\n== Testes: scoreBM25 + fusaoReciprocalRank (puros, sem rede) ==");

        List<String> corpusTeste = List.of(
                "idade mínima de doze anos para participar do estudo",
                "consentimento do responsável legal é obrigatório",
                "retirada do participante a qualquer momento sem justificativa");
        Bm25.EstatisticasBM25 estatisticasTeste = Bm25.construirEstatisticasBM25(corpusTeste);
        List<String> queryTeste = Tokenizer.tokenizar("qual a idade mínima exigida");
        double melhorScore = Double.NEGATIVE_INFINITY;
        int vencedorBM25 = -1;
        for (int idx = 0; idx < corpusTeste.size(); idx++) {
            double score = Bm25.scoreBM25(queryTeste, estatisticasTeste.tokensPorDoc().get(idx), estatisticasTeste);
            if (score > melhorScore) {
                melhorScore = score;
                vencedorBM25 = idx;
            }
        }
        total++;
        boolean okBM25 = vencedorBM25 == 0;
        System.out.printf("  [%s] scoreBM25: doc sobre idade mínima vence a busca lexical (venceu doc %d)%n",
                okBM25 ? "OK" : "FALHOU", vencedorBM25);
        if (okBM25) {
            passou++;
        }

        List<RankUtils.RankedScore> fusaoTeste = RankUtils.fusaoReciprocalRank(List.of(0, 1, 2), List.of(0, 2, 1), 60);
        total++;
        boolean okFusao = fusaoTeste.get(0).idx() == 0;
        System.out.printf("  [%s] fusaoReciprocalRank: doc no topo dos dois rankings vence a fusão (venceu doc %d)%n",
                okFusao ? "OK" : "FALHOU", fusaoTeste.get(0).idx());
        if (okFusao) {
            passou++;
        }

        System.out.printf("Total: %d teste(s), %d passou(passaram), %d falhou(falharam).%n%n", total, passou, total - passou);
        return passou == total;
    }
}
