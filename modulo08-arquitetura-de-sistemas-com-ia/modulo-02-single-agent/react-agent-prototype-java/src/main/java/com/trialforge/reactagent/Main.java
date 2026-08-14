package com.trialforge.reactagent;

/**
 * Protótipo: Agente único do TrialForge (geração da seção de assentimento do ICF).
 * Padrão: loop ReAct + ferramenta com schema tipado. Sem fila de mensagens.
 *
 * Modelo: Ollama local, rodando gemma4:e2b — gratuito, sem chave de API, roda inteiro
 * na máquina. Alternativas pagas (Claude, Gemini, GPT) estão em
 * {@link ProvedoresPagos} — mesmo loop ReAct, só a chamada ao modelo muda.
 *
 * Porte Java de react-agent-prototype.js / react_agent_prototype.py.
 * Ver README.md deste projeto para detalhes de paridade e adaptações.
 */
public class Main {

    private static final String MODELO = "gemma4:e2b";

    public static void main(String[] args) {
        try {
            rodarTestesFerramenta();

            String baseUrl = System.getenv().getOrDefault("OLLAMA_BASE_URL", "http://localhost:11434");
            ChatClient client = new RetryingChatClient(new OllamaReactClient(baseUrl, MODELO));

            Simulador.simularInteracao(client);
        } catch (Exception erro) {
            System.err.println("[Erro não tratado] " + erro.getMessage());
            System.out.println("[Sistema] Encaminhando ao Approval Gate — falha técnica também é motivo de escalonamento.");
        }
    }

    /**
     * Testes automatizados da ferramenta (determinístico, sem chamar o modelo).
     * Porte 1:1 de rodarTestesFerramenta em react-agent-prototype.js / .py.
     */
    static void rodarTestesFerramenta() {
        System.out.println("== Testes: executarBuscaClausula (determinístico, 6 casos + 1 de parâmetro inválido) ==");
        int passou = 0;

        for (CasosTesteFerramenta.Caso caso : CasosTesteFerramenta.CASOS) {
            ExecutorFerramenta.ResultadoBusca resultado = ExecutorFerramenta.executarBuscaClausula(
                    java.util.Map.of("tema", caso.tema(), "jurisdicao", caso.jurisdicao()));
            boolean achou = resultado.texto() != null;
            boolean ok = achou == caso.esperaAchar();
            System.out.println("  [" + (ok ? "OK" : "FALHOU") + "] tema=\"" + caso.tema() + "\", jurisdicao="
                    + caso.jurisdicao() + " -> achou=" + achou + " (esperado=" + caso.esperaAchar() + ")");
            if (ok) {
                passou++;
            }
        }

        // Reproduz o bug real já corrigido: parâmetro grafado errado ("jurisdicicao") não
        // pode estourar exceção pro chamador, tem que virar aviso de falha própria da ferramenta.
        ExecutorFerramenta.ResultadoBusca casoInvalido =
                ExecutorFerramenta.executarBuscaClausula(java.util.Map.of("tema", "x", "jurisdicicao", "ANVISA"));
        boolean invalidoOk = casoInvalido.texto() == null && casoInvalido.aviso() != null;
        System.out.println("  [" + (invalidoOk ? "OK" : "FALHOU")
                + "] parâmetro mal formado (\"jurisdicicao\") tratado como falha própria da ferramenta -> "
                + invalidoOk);
        if (invalidoOk) {
            passou++;
        }

        int total = CasosTesteFerramenta.CASOS.size() + 1;
        System.out.println("Total: " + total + " teste(s), " + passou + " passou(passaram), " + (total - passou)
                + " falhou(falharam).\n");
        if (passou != total) {
            throw new IllegalStateException(
                    "Testes da ferramenta falharam: " + passou + "/" + total + " — corrija antes de rodar a simulação.");
        }
    }
}
