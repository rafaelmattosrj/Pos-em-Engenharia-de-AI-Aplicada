package com.trialforge.gateway;

import java.text.Normalizer;
import java.util.Arrays;
import java.util.List;
import java.util.regex.Pattern;

/**
 * Normaliza para minusculas, decompoe acentos (NFD) e remove as marcas
 * combinantes resultantes, troca pontuacao por espaco e separa por espaco em
 * branco. Mesma receita usada nos dois originais para preparar tanto o
 * corpus quanto as queries do BM25 — idf mais estavel com corpus pequeno
 * quando os acentos nao geram tokens distintos por coincidencia.
 */
public final class Tokenizer {

    // \p{Mn} = marca combinante (Mark, Nonspacing) — o que sobra dos acentos
    // depois da decomposicao NFD (ex.: "í" -> "i" + U+0301).
    private static final Pattern MARCAS_COMBINANTES = Pattern.compile("\\p{Mn}");
    // \w em Java, por padrao, e ASCII ([a-zA-Z0-9_]) — mesmo comportamento do
    // \w em JS sem a flag "u" e do \w do modulo `re` do Python aplicado a um
    // texto ja reduzido a ASCII pela remocao de acentos acima.
    private static final Pattern NAO_PALAVRA = Pattern.compile("[^\\w\\s]");
    private static final Pattern ESPACOS = Pattern.compile("\\s+");

    private Tokenizer() {
    }

    public static List<String> tokenizar(String texto) {
        String minusculo = texto.toLowerCase();
        String decomposto = Normalizer.normalize(minusculo, Normalizer.Form.NFD);
        String semAcento = MARCAS_COMBINANTES.matcher(decomposto).replaceAll("");
        String semPontuacao = NAO_PALAVRA.matcher(semAcento).replaceAll(" ");
        String normalizado = semPontuacao.trim();
        if (normalizado.isEmpty()) {
            return List.of();
        }
        return Arrays.asList(ESPACOS.split(normalizado));
    }
}
