package com.finetuningapi;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.regex.Pattern;

/**
 * Deduplicacao via MinHash+LSH e balanceamento por amostragem com
 * temperatura + metricas de diversidade (entropia de Shannon / numero
 * efetivo de fontes).
 *
 * ADAPTACAO: este e um porte AUTOCONTIDO do algoritmo de
 * modulo-02-preparacao-datasets/dataset-cleaning-balancing-tool.js (mesma
 * formula, mesmos parametros MINHASH_K=32/semente=42/LSH bandas=8,linhas=4,
 * alpha=0.3, limiar de duplicata=0.55). O original em JS reusa o mesmo
 * arquivo via require() entre modulo-02 e modulo-03/09; como este repo nao
 * tem um mecanismo de modulo compartilhado entre projetos Maven/Go
 * independentes, a logica foi duplicada aqui deliberadamente (nao portada
 * por referencia), documentado tambem no README deste projeto.
 */
public final class MinHashDedupBalancer {

    public static final int MINHASH_K = 32;
    public static final long MINHASH_SEMENTE = 42;
    public static final int LSH_BANDAS = 8;
    public static final int LSH_LINHAS = 4;
    public static final double LIMIAR_DUPLICATA = 0.55;
    public static final double ALPHA_TEMPERATURA = 0.3;
    private static final BigInteger PRIMO_MERSENNE = BigInteger.valueOf(2147483647L);
    private static final Pattern ESPACOS = Pattern.compile("\\s+");

    private MinHashDedupBalancer() {
    }

    public static String normalizarTexto(String texto) {
        return ESPACOS.matcher(texto.toLowerCase()).replaceAll(" ").trim();
    }

    public static Set<String> shingles(String texto, int n) {
        String[] palavras = normalizarTexto(texto).split(" ");
        Set<String> conjunto = new HashSet<>();
        for (int i = 0; i <= palavras.length - n; i++) {
            conjunto.add(String.join(" ", List.of(palavras).subList(i, i + n)));
        }
        return conjunto;
    }

    public static double similaridadeJaccardExata(String a, String b, int n) {
        Set<String> sa = shingles(a, n);
        Set<String> sb = shingles(b, n);
        if (sa.isEmpty() || sb.isEmpty()) return 0;
        long intersecao = sa.stream().filter(sb::contains).count();
        long uniao = sa.size() + sb.size() - intersecao;
        return uniao == 0 ? 0 : (double) intersecao / uniao;
    }

    private static long hashString(String s) {
        long h = 5381;
        for (int i = 0; i < s.length(); i++) {
            h = ((h * 33) + s.charAt(i)) & 0xffffffffL;
        }
        return h;
    }

    public record CoeficienteHash(long a, long b) {
    }

    public static List<CoeficienteHash> gerarCoeficientesHash(int k, long semente) {
        long estado = semente & 0xffffffffL;
        List<CoeficienteHash> coeficientes = new ArrayList<>();
        for (int i = 0; i < k; i++) {
            estado = ((estado * 1103515245L) + 12345L) & 0xffffffffL;
            long a = (estado % 2000000000L) + 1;
            estado = ((estado * 1103515245L) + 12345L) & 0xffffffffL;
            long b = estado % 2000000000L;
            coeficientes.add(new CoeficienteHash(a, b));
        }
        return coeficientes;
    }

    private static long hashUniversal(long x, long a, long b) {
        return BigInteger.valueOf(a).multiply(BigInteger.valueOf(x)).add(BigInteger.valueOf(b))
                .mod(PRIMO_MERSENNE).longValue();
    }

    public static long[] assinaturaMinHash(Set<String> shingleSet, List<CoeficienteHash> coeficientes) {
        long[] baseHashes = shingleSet.stream().mapToLong(MinHashDedupBalancer::hashString).toArray();
        long[] assinatura = new long[coeficientes.size()];
        for (int i = 0; i < coeficientes.size(); i++) {
            CoeficienteHash c = coeficientes.get(i);
            long minimo = Long.MAX_VALUE;
            for (long x : baseHashes) {
                long h = hashUniversal(x, c.a(), c.b());
                if (h < minimo) minimo = h;
            }
            assinatura[i] = minimo;
        }
        return assinatura;
    }

    public static double similaridadeMinHashEstimada(long[] a, long[] b) {
        int iguais = 0;
        for (int i = 0; i < a.length; i++) if (a[i] == b[i]) iguais++;
        return (double) iguais / a.length;
    }

    public static Set<String> bandingLSH(List<long[]> assinaturas, int b, int r) {
        Map<String, List<Integer>> baldes = new HashMap<>();
        Set<String> candidatos = new HashSet<>();
        for (int idx = 0; idx < assinaturas.size(); idx++) {
            long[] assinatura = assinaturas.get(idx);
            for (int banda = 0; banda < b; banda++) {
                StringBuilder fatia = new StringBuilder();
                for (int k = banda * r; k < banda * r + r; k++) fatia.append(assinatura[k]).append(',');
                String chave = banda + ":" + fatia;
                List<Integer> balde = baldes.computeIfAbsent(chave, k -> new ArrayList<>());
                for (int outroIdx : balde) {
                    candidatos.add(outroIdx < idx ? outroIdx + "-" + idx : idx + "-" + outroIdx);
                }
                balde.add(idx);
            }
        }
        return candidatos;
    }

    public record Exemplo(String id, String caso, String fonte, String textoParaDedup) {
    }

    public record ParDuplicata(int i, int j, double similaridade) {
    }

    public record ResultadoDedup(List<ParDuplicata> paresDuplicata, long totalParesForcaBruta, long totalCandidatosLSH) {
    }

    /** Roda dedup restrito a um unico "caso" (grupo) de exemplos, generico -- nao hardcoded pra casos especificos. */
    public static ResultadoDedup encontrarQuaseDuplicatasGenerico(List<Exemplo> exemplos, int nShingle) {
        List<CoeficienteHash> coeficientes = gerarCoeficientesHash(MINHASH_K, MINHASH_SEMENTE);
        List<long[]> assinaturas = new ArrayList<>();
        for (Exemplo e : exemplos) assinaturas.add(assinaturaMinHash(shingles(e.textoParaDedup(), nShingle), coeficientes));
        Set<String> candidatos = bandingLSH(assinaturas, LSH_BANDAS, LSH_LINHAS);
        long totalForcaBruta = ((long) exemplos.size() * (exemplos.size() - 1)) / 2;

        List<ParDuplicata> pares = new ArrayList<>();
        for (String chave : candidatos) {
            String[] partes = chave.split("-");
            int li = Integer.parseInt(partes[0]);
            int lj = Integer.parseInt(partes[1]);
            double sim = similaridadeJaccardExata(exemplos.get(li).textoParaDedup(), exemplos.get(lj).textoParaDedup(), nShingle);
            if (sim >= LIMIAR_DUPLICATA) pares.add(new ParDuplicata(li, lj, sim));
        }
        return new ResultadoDedup(pares, totalForcaBruta, candidatos.size());
    }

    public static Map<String, Integer> contarPorFonte(List<Exemplo> exemplos, String caso) {
        Map<String, Integer> contagem = new LinkedHashMap<>();
        for (Exemplo e : exemplos) {
            if (e.caso().equals(caso)) contagem.merge(e.fonte(), 1, Integer::sum);
        }
        return contagem;
    }

    public static Map<String, Double> pesosAmostragemPorTemperatura(Map<String, Integer> contagens, double alpha) {
        Map<String, Double> pesosBrutos = new LinkedHashMap<>();
        double soma = 0;
        for (Map.Entry<String, Integer> entry : contagens.entrySet()) {
            double peso = Math.pow(entry.getValue(), alpha);
            pesosBrutos.put(entry.getKey(), peso);
            soma += peso;
        }
        Map<String, Double> pesos = new LinkedHashMap<>();
        double somaFinal = soma;
        pesosBrutos.forEach((fonte, peso) -> pesos.put(fonte, peso / somaFinal));
        return pesos;
    }

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
        List<Integer> ordemPorResto = new ArrayList<>();
        for (int i = 0; i < fontes.size(); i++) ordemPorResto.add(i);
        int alocadoBaseFinal = alocadoBase;
        ordemPorResto.sort((x, y) -> Double.compare(quotas[y] - base[y], quotas[x] - base[x]));

        Map<String, Integer> resultado = new LinkedHashMap<>();
        for (int i = 0; i < fontes.size(); i++) resultado.put(fontes.get(i), base[i]);
        int faltam = alvo - alocadoBaseFinal;
        for (int i = 0; i < faltam; i++) {
            String fonte = fontes.get(ordemPorResto.get(i));
            resultado.merge(fonte, 1, Integer::sum);
        }
        return resultado;
    }

    /** Alocacao capacitada: nunca aloca mais do que a fonte realmente tem disponivel. */
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
            for (String f : fontesAtivas) if (tentativa.get(f) > contagens.get(f)) excedentes.add(f);

            if (excedentes.isEmpty()) {
                fontesAtivas.forEach(f -> resultado.put(f, tentativa.get(f)));
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

    public static Map<String, Double> distribuicaoDe(Map<String, Integer> contagens) {
        int total = contagens.values().stream().mapToInt(Integer::intValue).sum();
        Map<String, Double> dist = new LinkedHashMap<>();
        contagens.forEach((fonte, n) -> dist.put(fonte, total == 0 ? 0 : (double) n / total));
        return dist;
    }

    public static double entropiaShannon(Map<String, Double> distribuicao) {
        return -distribuicao.values().stream().filter(p -> p > 0).mapToDouble(p -> p * Math.log(p)).sum();
    }

    public static double numeroEfetivoFontes(Map<String, Double> distribuicao) {
        return Math.exp(entropiaShannon(distribuicao));
    }
}
