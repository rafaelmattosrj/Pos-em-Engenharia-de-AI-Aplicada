package com.datasetprep.cleaning;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

/**
 * Porte da secao 2 de dataset-cleaning-balancing-tool.js: deduplicacao via
 * MinHash (Broder 1997) + LSH banding para geracao de candidatos, com
 * refinamento por similaridade de Jaccard exata so sobre os candidatos.
 * Mesma familia de tecnica do NEARDUP de Lee et al. 2021/2022 (arXiv 2107.06499).
 */
public final class MinHashLsh {

    private MinHashLsh() {
    }

    public static final double LIMIAR_DUPLICATA = 0.55;
    public static final int MINHASH_K = 32;
    public static final int MINHASH_SEMENTE = 42;
    public static final int LSH_BANDAS = 8;
    public static final int LSH_LINHAS = 4;

    private static final long PRIMO_MERSENNE = 2147483647L; // 2^31 - 1

    public static String normalizarTexto(String texto) {
        return texto.toLowerCase().replaceAll("\\s+", " ").trim();
    }

    public static Set<String> shingles(String texto, int n) {
        String[] palavras = normalizarTexto(texto).split(" ");
        Set<String> conjunto = new LinkedHashSet<>();
        for (int i = 0; i <= palavras.length - n; i++) {
            StringBuilder sb = new StringBuilder();
            for (int j = i; j < i + n; j++) {
                if (j > i) sb.append(' ');
                sb.append(palavras[j]);
            }
            conjunto.add(sb.toString());
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

    public static double similaridadeJaccardExata(String a, String b) {
        return similaridadeJaccardExata(a, b, 5);
    }

    /** djb2, hash de string deterministico, 32 bits sem sinal. */
    public static long hashString(String s) {
        long h = 5381;
        for (int i = 0; i < s.length(); i++) {
            h = ((h * 33) + s.charAt(i)) & 0xffffffffL;
        }
        return h;
    }

    public record CoeficienteHash(long a, long b) {
    }

    /** LCG deterministico: mesma semente sempre produz a mesma familia de funcoes hash. */
    public static List<CoeficienteHash> gerarCoeficientesHash(int k, long semente) {
        long[] estado = {semente & 0xffffffffL};
        List<CoeficienteHash> coeficientes = new ArrayList<>();
        for (int i = 0; i < k; i++) {
            estado[0] = ((estado[0] * 1103515245L) + 12345L) & 0xffffffffL;
            long a = (estado[0] % 2000000000L) + 1;
            estado[0] = ((estado[0] * 1103515245L) + 12345L) & 0xffffffffL;
            long b = estado[0] % 2000000000L;
            coeficientes.add(new CoeficienteHash(a, b));
        }
        return coeficientes;
    }

    public static long hashUniversal(long x, long a, long b) {
        return ((a * x) + b) % PRIMO_MERSENNE;
    }

    /** Assinatura MinHash: k valores minimos, um por funcao hash, sobre o conjunto de shingles. */
    public static long[] assinaturaMinHash(Set<String> shingleSet, List<CoeficienteHash> coeficientes) {
        long[] baseHashes = shingleSet.stream().mapToLong(MinHashLsh::hashString).toArray();
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

    /**
     * LSH banding: divide a assinatura de k valores em b bandas de r valores cada
     * (k = b*r). Dois exemplos viram "candidatos" se colidirem em pelo menos uma banda.
     */
    public static Set<String> bandingLSH(List<long[]> assinaturas, int b, int r) {
        Map<String, List<Integer>> baldes = new HashMap<>();
        Set<String> candidatos = new LinkedHashSet<>();
        for (int idx = 0; idx < assinaturas.size(); idx++) {
            long[] assinatura = assinaturas.get(idx);
            for (int banda = 0; banda < b; banda++) {
                StringBuilder fatia = new StringBuilder();
                for (int j = banda * r; j < banda * r + r; j++) {
                    if (j > banda * r) fatia.append(',');
                    fatia.append(assinatura[j]);
                }
                String chave = banda + ":" + fatia;
                List<Integer> balde = baldes.computeIfAbsent(chave, k -> new ArrayList<>());
                for (int outroIdx : balde) {
                    int menor = Math.min(outroIdx, idx);
                    int maior = Math.max(outroIdx, idx);
                    candidatos.add(menor + "-" + maior);
                }
                balde.add(idx);
            }
        }
        return candidatos;
    }

    public record ParDuplicata(int i, int j, double similaridade) {
    }

    public record ResultadoPorCaso(
            int itensNoCaso, long paresForcaBruta, int candidatosLSH, int duplicatasConfirmadas, double reducaoPercentual
    ) {
    }

    public record ResultadoDedup(
            List<ParDuplicata> paresDuplicata,
            Map<String, ResultadoPorCaso> resultadosPorCaso,
            long totalParesForcaBruta,
            long totalCandidatosLSH
    ) {
    }

    public static ResultadoDedup encontrarQuaseDuplicatasMinHashLSH(List<Exemplo> exemplos) {
        List<CoeficienteHash> coeficientes = gerarCoeficientesHash(MINHASH_K, MINHASH_SEMENTE);
        Map<String, ResultadoPorCaso> resultadosPorCaso = new HashMap<>();
        long totalParesForcaBruta = 0;
        long totalCandidatosLSH = 0;
        List<ParDuplicata> paresDuplicata = new ArrayList<>();

        for (String caso : List.of("amplitude-auto", "amplitude-saude-empresarial")) {
            List<Integer> indicesGlobais = new ArrayList<>();
            List<Exemplo> itens = new ArrayList<>();
            for (int i = 0; i < exemplos.size(); i++) {
                if (exemplos.get(i).metadata().caso().equals(caso)) {
                    indicesGlobais.add(i);
                    itens.add(exemplos.get(i));
                }
            }
            List<long[]> assinaturas = new ArrayList<>();
            for (Exemplo e : itens) {
                assinaturas.add(assinaturaMinHash(shingles(e.entrada(), 5), coeficientes));
            }
            Set<String> candidatosLocais = bandingLSH(assinaturas, LSH_BANDAS, LSH_LINHAS);
            long paresForcaBruta = (long) itens.size() * (itens.size() - 1) / 2;

            int confirmados = 0;
            for (String chave : candidatosLocais) {
                String[] partes = chave.split("-");
                int li = Integer.parseInt(partes[0]);
                int lj = Integer.parseInt(partes[1]);
                double simExata = similaridadeJaccardExata(itens.get(li).entrada(), itens.get(lj).entrada());
                if (simExata >= LIMIAR_DUPLICATA) {
                    confirmados++;
                    paresDuplicata.add(new ParDuplicata(indicesGlobais.get(li), indicesGlobais.get(lj), simExata));
                }
            }

            double reducaoPercentual = paresForcaBruta == 0 ? 0 : 100 * (1 - (double) candidatosLocais.size() / paresForcaBruta);
            resultadosPorCaso.put(caso, new ResultadoPorCaso(itens.size(), paresForcaBruta, candidatosLocais.size(), confirmados, reducaoPercentual));
            totalParesForcaBruta += paresForcaBruta;
            totalCandidatosLSH += candidatosLocais.size();
        }

        return new ResultadoDedup(paresDuplicata, resultadosPorCaso, totalParesForcaBruta, totalCandidatosLSH);
    }

    public record ResultadoRemocao(
            List<Exemplo> mantidos,
            List<ParDuplicata> paresDuplicata,
            Map<String, ResultadoPorCaso> resultadosPorCaso,
            long totalParesForcaBruta,
            long totalCandidatosLSH,
            int removidos
    ) {
    }

    /** Mantem a primeira ocorrencia de cada par de quase-duplicata, remove a segunda. */
    public static ResultadoRemocao removerQuaseDuplicatas(List<Exemplo> exemplos) {
        ResultadoDedup dedup = encontrarQuaseDuplicatasMinHashLSH(exemplos);
        Set<Integer> remover = new LinkedHashSet<>();
        for (ParDuplicata p : dedup.paresDuplicata()) remover.add(p.j());

        List<Exemplo> mantidos = new ArrayList<>();
        for (int i = 0; i < exemplos.size(); i++) {
            if (!remover.contains(i)) mantidos.add(exemplos.get(i));
        }
        return new ResultadoRemocao(mantidos, dedup.paresDuplicata(), dedup.resultadosPorCaso(),
                dedup.totalParesForcaBruta(), dedup.totalCandidatosLSH(), remover.size());
    }
}
