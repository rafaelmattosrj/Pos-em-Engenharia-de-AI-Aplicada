package com.trialforge.agentcomponents;

import java.io.IOException;
import java.util.List;

/**
 * 2. PLANEJAMENTO — chain-of-thought (quase de graca) vs reflexao (chamada a mais).
 *
 * Unica secao que chama o modelo de verdade (Ollama) — por isso, diferente de
 * Memoria/Ferramentas/AcaoGate, nao tem teste automatizado equivalente no
 * original (so observacao ao vivo em demonstrarPlanejamento()).
 *
 * Porte 1:1 de agent-components-demo.js / agent_components_demo.py (secao 2).
 */
public final class Planejamento {

    private Planejamento() {
    }

    public record ResultadoPlanejamento(String texto, long duracaoMs, int chamadas) {
    }

    public static ResultadoPlanejamento chainOfThought(ChatClient client, String pergunta)
            throws IOException, InterruptedException {
        long inicio = System.currentTimeMillis();
        String resposta = client.chat(List.of(
                new Memoria.Mensagem("user", "Pense passo a passo, de forma breve, antes de responder: " + pergunta)
        ));
        return new ResultadoPlanejamento(resposta, System.currentTimeMillis() - inicio, 1);
    }

    public static ResultadoPlanejamento chainOfThoughtMaisReflexao(ChatClient client, String pergunta)
            throws IOException, InterruptedException {
        long inicio = System.currentTimeMillis();
        String primeira = client.chat(List.of(new Memoria.Mensagem("user", pergunta)));
        String critica = client.chat(List.of(
                new Memoria.Mensagem("user", pergunta),
                new Memoria.Mensagem("assistant", primeira),
                new Memoria.Mensagem("user",
                        "Releia sua resposta acima. Ela tem algum erro ou imprecisão? Responda só \"sim\" ou \"não\" e, se sim, qual.")
        ));
        return new ResultadoPlanejamento(critica, System.currentTimeMillis() - inicio, 2);
    }
}
