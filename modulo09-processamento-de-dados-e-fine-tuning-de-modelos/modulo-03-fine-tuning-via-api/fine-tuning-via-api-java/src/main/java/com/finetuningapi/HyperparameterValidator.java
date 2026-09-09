package com.finetuningapi;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Validacao client-side de hiperparametro ANTES do envio (Modulo 3.3/3.4):
 * a Vertex AI aceita epochCount=0 sem erro rapido e substitui por um default
 * silencioso, entao a validacao real precisa acontecer aqui, nao la.
 * Tambem compara o hiperparametro pedido com o que a API realmente aplicou
 * (campos int64 voltam como String no JSON da API, entao a comparacao
 * tolera esse tipo diferente).
 */
public final class HyperparameterValidator {

    public record FaixaValida(double min, double max) {
        boolean contem(double valor) {
            return valor >= min && valor <= max;
        }
    }

    public static final FaixaValida EPOCH_COUNT = new FaixaValida(1, 20);
    public static final FaixaValida LEARNING_RATE_MULTIPLIER = new FaixaValida(0.1, 10);

    public record Hiperparametros(Integer epochCount, Double learningRateMultiplier) {
    }

    private HyperparameterValidator() {
    }

    public static void validar(Hiperparametros config) {
        List<String> erros = new ArrayList<>();
        if (config.epochCount() == null || !EPOCH_COUNT.contem(config.epochCount())) {
            erros.add("epochCount deve ser inteiro entre " + (int) EPOCH_COUNT.min() + " e " + (int) EPOCH_COUNT.max()
                    + ", recebido: " + config.epochCount());
        }
        if (config.learningRateMultiplier() == null || !LEARNING_RATE_MULTIPLIER.contem(config.learningRateMultiplier())) {
            erros.add("learningRateMultiplier deve estar entre " + LEARNING_RATE_MULTIPLIER.min() + " e "
                    + LEARNING_RATE_MULTIPLIER.max() + ", recebido: " + config.learningRateMultiplier());
        }
        if (!erros.isEmpty()) {
            throw new IllegalArgumentException("Hiperparâmetro inválido, job não enviado:\n  " + String.join("\n  ", erros));
        }
    }

    /**
     * Compara dois valores tolerando a diferenca de tipo que a Vertex AI
     * introduz de verdade: campos int64 (epochCount) voltam como String no
     * JSON ("3"), nao como numero. Se ambos forem numericamente comparaveis,
     * compara como numero; senao cai pra comparacao direta de String.
     */
    public static boolean valoresEquivalentes(Object a, Object b) {
        Double na = tentarNumero(a);
        Double nb = tentarNumero(b);
        if (na != null && nb != null) return na.doubleValue() == nb.doubleValue();
        return String.valueOf(a).equals(String.valueOf(b));
    }

    private static Double tentarNumero(Object valor) {
        if (valor == null) return null;
        try {
            return Double.parseDouble(String.valueOf(valor));
        } catch (NumberFormatException e) {
            return null;
        }
    }

    public static List<String> compararHiperparametros(Map<String, Object> pedido, Map<String, Object> aplicado) {
        List<String> divergencias = new ArrayList<>();
        for (Map.Entry<String, Object> entry : pedido.entrySet()) {
            String chave = entry.getKey();
            Object valorPedido = entry.getValue();
            if (!aplicado.containsKey(chave) || aplicado.get(chave) == null) {
                divergencias.add(chave + ": pedido " + valorPedido + ", aplicado ausente (provavelmente default silencioso do provedor)");
            } else if (!valoresEquivalentes(aplicado.get(chave), valorPedido)) {
                divergencias.add(chave + ": pedido " + valorPedido + ", aplicado " + aplicado.get(chave));
            }
        }
        return divergencias;
    }

    public static Map<String, Object> mapaHiperparametros(Object epochCount) {
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("epochCount", epochCount);
        return m;
    }
}
