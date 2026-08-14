package com.trialforge.tiering;

import java.nio.file.Path;
import java.util.List;
import java.util.concurrent.Callable;

/**
 * Porte Java de trialforge-model-tiering-prototype.js / trialforge_model_tiering_prototype.py.
 * Ver README.md deste projeto para detalhes de paridade e adaptações.
 *
 * {@code mvn exec:java} roda a demo gravada (4 chamadas sequenciais).
 * {@code mvn exec:java -Dexec.args="--volume"} roda o extra de volume concorrente
 * (Missão Prática) — não precisa de aprovação humana (sem CSR na mistura).
 */
public class Main {

    public static void main(String[] args) {
        String baseUrl = System.getenv().getOrDefault("OLLAMA_BASE_URL", "http://localhost:11434");
        OllamaGateway ollama = new OllamaHttpGateway(baseUrl);
        RagIndex ragIndex = new RagIndex(ollama, CascadeGateway.MODELO_EMBEDDING);
        OrcamentoManager orcamento = new OrcamentoManager();
        AuditTrail auditTrail = new AuditTrail(Path.of("audit-trail-tiering.jsonl"));

        boolean rodarVolume = args.length > 0 && "--volume".equals(args[0]);

        try {
            ragIndex.prepararIndice();

            if (rodarVolume) {
                rodarVolume(ollama, ragIndex, orcamento, auditTrail);
            } else {
                orcamento.definirOrcamento("estudo-A", 0.05);
                orcamento.definirOrcamento("estudo-B", 0.005);
                CascadeGateway gateway = new CascadeGateway(ollama, ragIndex, orcamento, new StdinApprovalPrompt(), auditTrail);
                rodarDemo(gateway);
                AuditTrailVerifier.verificarEImprimir(auditTrail);
            }
        } catch (Exception erro) {
            System.err.println("[Erro não tratado] " + erro.getMessage());
            System.exit(1);
        }
    }

    private static void rodarDemo(CascadeGateway gateway) throws Exception {
        // 1) Estudo A, pergunta que o Tier 1 resolve bem sozinho
        gateway.processarComCascata("Quais são as regras de assentimento pra menores nesse estudo?", "estudo-A");

        // 2) Estudo A, pergunta fora do assunto do banco de cláusulas — escala pro Tier 2
        gateway.processarComCascata("Qual é o prazo de validade dos exames laboratoriais desse estudo?", "estudo-A");

        // 3) Estudo A de novo, mas agora síntese de CSR — regra fixa, direto pro Tier 2
        gateway.processarComCascata("Preciso da síntese do CSR final desse estudo.", "estudo-A");

        // 4) Estudo B, com orçamento propositalmente baixo — deve bloquear antes de chamar qualquer modelo
        gateway.processarComCascata("Quais são as regras de assentimento pra menores nesse estudo?", "estudo-B");
    }

    private static void rodarVolume(OllamaGateway ollama, RagIndex ragIndex, OrcamentoManager orcamento, AuditTrail auditTrail)
            throws InterruptedException {
        final String perguntaAssentimento = "Quais são as regras de assentimento pra menores nesse estudo?";
        final String perguntaRetirada = "O participante pode desistir do estudo a qualquer momento?";
        final String perguntaForaDominio = "Qual é o prazo de validade dos exames laboratoriais desse estudo?";

        orcamento.definirOrcamento("estudo-C", 0.05);
        orcamento.definirOrcamento("estudo-D", 0.05);
        orcamento.definirOrcamento("estudo-E", 0.05);
        // Orçamento apertado de propósito: reservarOrcamento reserva o PIOR caso por requisição
        // (CUSTO_TIER1+CUSTO_TIER2 = 0.011) — esse limite cabe exatamente 2 reservas, com 5
        // requisições concorrentes disputando o mesmo estudo.
        orcamento.definirOrcamento("estudo-F", 2 * (CascadeGateway.CUSTO_TIER1 + CascadeGateway.CUSTO_TIER2) + 0.0005);

        // Approval Gate não é acionado (sem CSR na mistura), então um ApprovalPrompt que
        // sempre aprova nunca chega a ser usado de fato.
        CascadeGateway gateway = new CascadeGateway(ollama, ragIndex, orcamento, rascunho -> true, auditTrail);

        List<Callable<String>> requisicoes = new java.util.ArrayList<>();
        for (String estudoId : List.of("estudo-C", "estudo-D", "estudo-E")) {
            requisicoes.add(() -> gateway.processarComCascata(perguntaAssentimento, estudoId));
            requisicoes.add(() -> gateway.processarComCascata(perguntaRetirada, estudoId));
            requisicoes.add(() -> gateway.processarComCascata(perguntaAssentimento, estudoId));
            requisicoes.add(() -> gateway.processarComCascata(perguntaForaDominio, estudoId));
        }
        for (int i = 0; i < 5; i++) {
            requisicoes.add(() -> gateway.processarComCascata(perguntaAssentimento, "estudo-F"));
        }

        System.out.printf("%n== Volume concorrente: %d requisições, 4 estudos, disparadas ao mesmo tempo ==%n%n", requisicoes.size());
        List<VolumeSimulator.ResultadoEstudo> resultados = VolumeSimulator.simular(
                gateway, orcamento, List.of("estudo-C", "estudo-D", "estudo-E", "estudo-F"), requisicoes, 20);

        System.out.println("\n== Resultado por estudo, depois da concorrência ==");
        boolean algumEstourou = false;
        for (VolumeSimulator.ResultadoEstudo resultado : resultados) {
            if (resultado.estourou()) algumEstourou = true;
            System.out.printf("  [%s] %s: gasto %.4f / limite %.4f%n",
                    resultado.estourou() ? "ESTOUROU" : "OK", resultado.estudoId(), resultado.gasto(), resultado.limite());
        }
        if (algumEstourou) {
            throw new IllegalStateException("Volume concorrente estourou orçamento de pelo menos um estudo — corrija reservarOrcamento.");
        }
        System.out.println("\n[OK] Nenhum estudo gastou além do limite, mesmo com requisições simultâneas pro mesmo estudo.");
    }
}
