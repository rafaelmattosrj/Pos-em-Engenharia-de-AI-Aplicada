package com.trialforge.agentcomponents;

import java.util.List;

/**
 * Cinco mini-demonstracoes, uma por peca da anatomia de um agente unico
 * (Memoria, Planejamento, Ferramentas, Acao, Approval Gate) — cada uma isolada
 * e rodavel sozinha, sem loop ReAct inteiro (isso e react-agent-prototype) e
 * sem schema formal de ferramenta.
 *
 * Contexto: TrialForge, Agente ICF (gera a secao de assentimento do Termo de
 * Consentimento a partir do protocolo do estudo).
 *
 * Porte 1:1 de agent-components-demo.js / agent_components_demo.py.
 * Ver README.md deste projeto para detalhes de paridade e adaptacoes.
 */
public class Main {

    private static final String MODELO = "gemma4:e2b";

    public static void main(String[] args) {
        rodarTestes();
        demonstrarMemoria();
        demonstrarPlanejamento();
        demonstrarFerramentas();
        demonstrarAcaoEGate();
    }

    // ============================================================
    // Testes automatizados (mesma disciplina do original: so a parte
    // deterministica ganha assert; Planejamento fica pra observacao ao vivo)
    // ============================================================

    private static void rodarTestes() {
        testarMemoria();
        testarFerramenta();
        testarGate();
    }

    private static void testarMemoria() {
        System.out.println("== Testes: memória de longo prazo persiste entre chamadas ==");
        Memoria.MemoriaLongoPrazo banco = new Memoria.MemoriaLongoPrazo();
        Memoria.Historico primeira = banco.registrar("usuario-teste", "prefere respostas curtas");
        Memoria.Historico segunda = banco.registrar("usuario-teste", "prefere tom formal");
        boolean ok = segunda.interacoes() == 2 && segunda.preferencias().size() == 2 && primeira.interacoes() == 1;
        System.out.println("  [" + (ok ? "OK" : "FALHOU") + "] duas chamadas com o mesmo usuarioId acumulam estado");
        if (!ok) {
            throw new IllegalStateException("memoriaLongoPrazo não está persistindo corretamente entre chamadas");
        }
        System.out.println();
    }

    private static void testarFerramenta() {
        System.out.println("== Testes: buscarClausulaAssentimento (determinístico) ==");
        Ferramentas.ResultadoClausula menor = Ferramentas.buscarClausulaAssentimento(List.of(12, 15, 17));
        Ferramentas.ResultadoClausula adulto = Ferramentas.buscarClausulaAssentimento(List.of(25, 40, 55));
        boolean okMenor = menor.texto() != null;
        boolean okAdulto = adulto.texto() == null;
        System.out.println("  [" + (okMenor ? "OK" : "FALHOU") + "] faixa com menor de idade encontra cláusula");
        System.out.println("  [" + (okAdulto ? "OK" : "FALHOU") + "] faixa só de adultos não encontra cláusula");
        if (!okMenor || !okAdulto) {
            throw new IllegalStateException("buscarClausulaAssentimento não está classificando corretamente");
        }
        System.out.println();
    }

    private static void testarGate() {
        System.out.println("== Testes: executarOuGatear ==");
        AcaoGate.ResultadoAcao semGate =
                AcaoGate.executarOuGatear(new AcaoGate.AcaoProposta("x", false, () -> "feito"));
        AcaoGate.ResultadoAcao comGate =
                AcaoGate.executarOuGatear(new AcaoGate.AcaoProposta("y", true, () -> "nunca deveria rodar"));
        boolean okSemGate = semGate.status().equals("executada") && "feito".equals(semGate.resultado());
        boolean okComGate = comGate.status().equals("aguardando_aprovacao") && comGate.resultado() == null;
        System.out.println("  [" + (okSemGate ? "OK" : "FALHOU") + "] ação sem gate executa direto");
        System.out.println("  [" + (okComGate ? "OK" : "FALHOU") + "] ação com gate nunca chega a executar");
        if (!okSemGate || !okComGate) {
            throw new IllegalStateException("executarOuGatear não está bloqueando/liberando corretamente");
        }
        System.out.println();
    }

    // ============================================================
    // Demonstracoes (saida em console, replicando o original)
    // ============================================================

    private static void demonstrarMemoria() {
        System.out.println("===== 1. MEMÓRIA =====");

        List<Memoria.Mensagem> contexto = Memoria.memoriaCurtoPrazo(
                "Estudo fase II, público-alvo 12-17 anos, terapia oncológica experimental.");
        System.out.println(
                "[Curto prazo] Contexto construído para esta requisição: " + contexto.size() + " mensagem(ns).");
        System.out.println("  -> O Agente ICF do TrialForge só precisa disso: cada protocolo é um caso novo,");
        System.out.println("     memória de longo prazo aqui seria custo sem benefício (canvas do Módulo 2.1).");
        System.out.println();

        System.out.println(
                "[Longo prazo] Simulando um assistente diferente, que acompanha o mesmo usuário ao longo do tempo:");
        Memoria.MemoriaLongoPrazo banco = new Memoria.MemoriaLongoPrazo();
        System.out.println("  1ª interação: " + banco.registrar("usuario-42", "prefere respostas curtas"));
        System.out.println("  2ª interação: " + banco.registrar("usuario-42", "prefere tom formal"));
        System.out.println("  -> Repare: a segunda chamada já sabe da primeira. Isso só vale a pena quando");
        System.out.println("     a tarefa se repete com o mesmo contexto ao longo do tempo — o oposto do ICF.");
        System.out.println();
    }

    private static void demonstrarPlanejamento() {
        System.out.println("===== 2. PLANEJAMENTO =====");
        String pergunta = "Um estudo com público-alvo de 12 a 17 anos precisa de assentimento do participante, "
                + "além do consentimento do responsável?";

        String baseUrl = System.getenv().getOrDefault("OLLAMA_BASE_URL", "http://localhost:11434");
        ChatClient client = new OllamaChatClient(baseUrl, MODELO);

        try {
            Planejamento.ResultadoPlanejamento cot = Planejamento.chainOfThought(client, pergunta);
            System.out.println(
                    "[Chain-of-thought] " + cot.chamadas() + " chamada ao modelo, " + cot.duracaoMs() + "ms.");

            Planejamento.ResultadoPlanejamento reflexao = Planejamento.chainOfThoughtMaisReflexao(client, pergunta);
            System.out.println("[+ Reflexão]        " + reflexao.chamadas() + " chamadas ao modelo, "
                    + reflexao.duracaoMs() + "ms.");

            double razao = cot.duracaoMs() == 0 ? 0 : (double) reflexao.duracaoMs() / cot.duracaoMs();
            System.out.printf("  -> Reflexão levou %.1fx mais tempo que chain-of-thought sozinho,%n", razao);
            System.out.println(
                    "     porque é uma chamada inteira a mais, não só um raciocínio mais longo na mesma chamada.");
            System.out.println("     Essa é a distinção de custo que o canvas do Módulo 2.1 registra.");
        } catch (Exception erro) {
            System.out.println("[Erro] Não foi possível chamar o modelo: " + erro.getMessage());
            System.out.println(
                    "Verifique se o Ollama está rodando ('ollama serve') e o modelo baixado ('ollama pull "
                            + MODELO + "').");
        }
        System.out.println();
    }

    private static void demonstrarFerramentas() {
        System.out.println("===== 3. FERRAMENTAS =====");
        System.out.println("[Caso 1] Estudo com participantes de 12 a 17 anos:");
        System.out.println("   " + Ferramentas.buscarClausulaAssentimento(List.of(12, 15, 17)));
        System.out.println("[Caso 2] Estudo só com adultos:");
        System.out.println("   " + Ferramentas.buscarClausulaAssentimento(List.of(25, 40, 55)));
        System.out.println("  -> O agente não \"sabe\" essa regra de cor: ele delega pra uma função");
        System.out.println("     determinística, porque é isso que sistemas determinísticos fazem melhor.");
        System.out.println();
    }

    private static void demonstrarAcaoEGate() {
        System.out.println("===== 4 e 5. AÇÃO + APPROVAL GATE =====");

        AcaoGate.AcaoProposta gerarRascunho = new AcaoGate.AcaoProposta(
                "gerar_rascunho_icf", false, () -> "Rascunho da seção de assentimento gerado.");
        System.out.println("[Ação 1: gerar rascunho]        " + AcaoGate.executarOuGatear(gerarRascunho));

        AcaoGate.AcaoProposta notificarEventoAdverso = new AcaoGate.AcaoProposta(
                "notificar_evento_adverso_regulatorio", true, () -> "Notificação enviada à ANVISA.");
        System.out.println(
                "[Ação 2: notificar evento adverso] " + AcaoGate.executarOuGatear(notificarEventoAdverso));

        System.out.println("  -> Mesma \"capacidade de ação\" nos dois casos. A diferença que decide se");
        System.out.println("     executa direto ou pausa pro Approval Gate não é técnica, é a Pergunta 2");
        System.out.println("     do Módulo 1.3: o erro é caro E irreversível?");
    }
}
