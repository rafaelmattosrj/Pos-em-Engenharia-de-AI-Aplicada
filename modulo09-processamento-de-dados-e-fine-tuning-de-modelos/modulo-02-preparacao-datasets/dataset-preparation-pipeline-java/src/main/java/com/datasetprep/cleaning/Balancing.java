package com.datasetprep.cleaning;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Porte da secao 3 de dataset-cleaning-balancing-tool.js: amostragem por
 * temperatura (Raffel et al. 2020, T5; Xue et al. 2021, mT5, alpha=0.3) +
 * alocacao capacitada pelo metodo do maior resto (Hamilton/Hare-Niemeyer).
 */
public final class Balancing {

    private Balancing() {
    }

    public static final double ALPHA_TEMPERATURA = 0.3;

    public static Map<String, Integer> contarPorFonte(List<Exemplo> exemplos, String caso) {
        Map<String, Integer> contagem = new LinkedHashMap<>();
        for (Exemplo e : exemplos) {
            if (e.metadata().caso().equals(caso)) {
                contagem.merge(e.metadata().fonte(), 1, Integer::sum);
            }
        }
        return contagem;
    }

    /** Peso de amostragem por fonte: p_i proporcional a n_i^alpha. */
    public static Map<String, Double> pesosAmostragemPorTemperatura(Map<String, Integer> contagens, double alpha) {
        Map<String, Double> pesosBrutos = new LinkedHashMap<>();
        double soma = 0;
        for (Map.Entry<String, Integer> e : contagens.entrySet()) {
            double p = Math.pow(e.getValue(), alpha);
            pesosBrutos.put(e.getKey(), p);
            soma += p;
        }
        Map<String, Double> pesos = new LinkedHashMap<>();
        for (Map.Entry<String, Double> e : pesosBrutos.entrySet()) {
            pesos.put(e.getKey(), e.getValue() / soma);
        }
        return pesos;
    }

    /** Metodo do maior resto (Hamilton/Hare-Niemeyer): converte pesos continuos em contagem inteira que soma exatamente ao alvo. */
    public static Map<String, Integer> alocarMaiorResto(Map<String, Double> pesos, int alvo) {
        List<String> fontes = new ArrayList<>(pesos.keySet());
        double[] quotas = new double[fontes.size()];
        int[] base = new int[fontes.size()];
        int alocadoBase = 0;
        for (int i = 0; i < fontes.size(); i++) {
            quotas[i] = pesos.get(fontes.get(i)) * alvo;
            base[i] = (int) Math.floor(quotas[i]);
            alocadoBase += base[i];
        }

        record Resto(String fonte, double resto) {
        }
        List<Resto> restos = new ArrayList<>();
        for (int i = 0; i < fontes.size(); i++) {
            restos.add(new Resto(fontes.get(i), quotas[i] - base[i]));
        }
        restos.sort(Comparator.comparingDouble(Resto::resto).reversed());

        int faltam = alvo - alocadoBase;
        Map<String, Integer> resultado = new LinkedHashMap<>();
        for (int i = 0; i < fontes.size(); i++) resultado.put(fontes.get(i), base[i]);
        for (int i = 0; i < faltam; i++) {
            String f = restos.get(i).fonte();
            resultado.merge(f, 1, Integer::sum);
        }
        return resultado;
    }

    /**
     * Alocacao capacitada: aplica o peso por temperatura, mas nunca aloca mais
     * do que a fonte realmente tem disponivel. Fontes que estourariam a
     * capacidade sao fixadas no maximo disponivel, e o alvo restante e
     * redistribuido por temperatura entre as fontes que sobraram - processo
     * iterativo ate estabilizar.
     */
    public static Map<String, Integer> alocarComCapacidade(Map<String, Integer> contagens, double alpha, int alvoTotal) {
        List<String> fontesAtivas = new ArrayList<>(contagens.keySet());
        int alvoRestante = alvoTotal;
        Map<String, Integer> resultado = new LinkedHashMap<>();
        int maxIteracoes = contagens.size() + 1;

        for (int iter = 0; iter < maxIteracoes && !fontesAtivas.isEmpty(); iter++) {
            Map<String, Integer> contagensAtivas = new LinkedHashMap<>();
            for (String f : fontesAtivas) contagensAtivas.put(f, contagens.get(f));
            Map<String, Double> pesos = pesosAmostragemPorTemperatura(contagensAtivas, alpha);
            Map<String, Integer> tentativa = alocarMaiorResto(pesos, alvoRestante);

            List<String> excedentes = new ArrayList<>();
            for (String f : fontesAtivas) {
                if (tentativa.get(f) > contagens.get(f)) excedentes.add(f);
            }
            if (excedentes.isEmpty()) {
                for (String f : fontesAtivas) resultado.put(f, tentativa.get(f));
                break;
            }
            for (String f : excedentes) {
                resultado.put(f, contagens.get(f));
                alvoRestante -= contagens.get(f);
            }
            fontesAtivas.removeAll(excedentes);
        }
        return resultado;
    }

    public record ResultadoBalanceamento(List<Exemplo> exemplos, Map<String, Integer> alocacao, Map<String, Integer> contagens) {
    }

    public static ResultadoBalanceamento balancearPorTemperatura(List<Exemplo> exemplos, String caso, double alpha, int alvoTotal) {
        Map<String, Integer> contagens = contarPorFonte(exemplos, caso);
        Map<String, Integer> alocacao = alocarComCapacidade(contagens, alpha, alvoTotal);

        List<Exemplo> doCaso = new ArrayList<>();
        List<Exemplo> outros = new ArrayList<>();
        for (Exemplo e : exemplos) {
            if (e.metadata().caso().equals(caso)) doCaso.add(e);
            else outros.add(e);
        }

        Map<String, Integer> contadorUsado = new LinkedHashMap<>();
        List<Exemplo> selecionados = new ArrayList<>();
        for (Exemplo e : doCaso) {
            String fonte = e.metadata().fonte();
            int usado = contadorUsado.getOrDefault(fonte, 0);
            if (usado < alocacao.getOrDefault(fonte, 0)) {
                selecionados.add(e);
                contadorUsado.put(fonte, usado + 1);
            }
        }

        List<Exemplo> resultado = new ArrayList<>(outros);
        resultado.addAll(selecionados);
        return new ResultadoBalanceamento(resultado, alocacao, contagens);
    }
}
