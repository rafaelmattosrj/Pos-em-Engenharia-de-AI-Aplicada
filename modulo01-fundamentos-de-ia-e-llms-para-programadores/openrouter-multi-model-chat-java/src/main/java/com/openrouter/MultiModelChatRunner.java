package com.openrouter;

import java.util.ArrayList;
import java.util.List;

/**
 * Envia a mesma pergunta para uma lista de modelos do OpenRouter, capturando
 * sucesso/erro de cada chamada individualmente — uma falha em um modelo nao
 * interrompe os demais.
 *
 * Extraido de Main.java para tornar a orquestracao testavel sem depender de
 * chamadas HTTP reais (equivalente ao pacote Go `openrouter`, testado via
 * servidor HTTP em memoria em client_test.go).
 */
public class MultiModelChatRunner {

    /**
     * Resultado de uma chamada a um modelo: {@code erro == null} indica
     * sucesso, com a resposta preenchida em {@code resposta}; caso
     * contrario, {@code erro} contem a mensagem de falha.
     */
    public record Resultado(String modelo, String resposta, String erro) {
        public boolean sucesso() {
            return erro == null;
        }
    }

    private final ModeloInvoker invoker;

    public MultiModelChatRunner(ModeloInvoker invoker) {
        this.invoker = invoker;
    }

    public List<Resultado> executar(List<String> modelos, String mensagem) {
        List<Resultado> resultados = new ArrayList<>();

        for (String modelo : modelos) {
            try {
                String resposta = invoker.complete(modelo, mensagem);
                resultados.add(new Resultado(modelo, resposta, null));
            } catch (Exception e) {
                resultados.add(new Resultado(modelo, null, e.getMessage()));
            }
        }

        return resultados;
    }
}
