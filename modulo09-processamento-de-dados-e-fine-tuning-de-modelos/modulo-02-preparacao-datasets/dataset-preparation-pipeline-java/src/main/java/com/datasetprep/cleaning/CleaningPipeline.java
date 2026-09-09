package com.datasetprep.cleaning;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/** Porte da secao 5 de dataset-cleaning-balancing-tool.js: pipeline completo (dedup -> balanceamento -> diversidade). */
public final class CleaningPipeline {

    private CleaningPipeline() {
    }

    public record RelatorioPorCaso(
            Map<String, Integer> contagensAntes, Map<String, Double> distAntes, double entropiaAntes, double nEfetivoAntes,
            Map<String, Integer> contagensDepois, Map<String, Double> distDepois, double entropiaDepois, double nEfetivoDepois,
            Map<String, Integer> alocacao
    ) {
    }

    public record ResultadoPipeline(
            int original, int aposDedup, int duplicatasRemovidas,
            Map<String, MinHashLsh.ResultadoPorCaso> resultadosDedupPorCaso,
            long totalParesForcaBruta, long totalCandidatosLSH,
            int fin, Map<String, RelatorioPorCaso> relatorioPorCaso, List<Exemplo> exemplosFinal
    ) {
    }

    public static ResultadoPipeline limparEBalancear(List<Exemplo> exemplos, double alpha, Map<String, Integer> alvos) {
        MinHashLsh.ResultadoRemocao dedup = MinHashLsh.removerQuaseDuplicatas(exemplos);
        List<Exemplo> atual = dedup.mantidos();
        Map<String, RelatorioPorCaso> relatorioPorCaso = new LinkedHashMap<>();

        for (String caso : List.of("amplitude-auto", "amplitude-saude-empresarial")) {
            Map<String, Integer> contagensAntes = Balancing.contarPorFonte(atual, caso);
            Map<String, Double> distAntes = Diversity.distribuicaoDe(contagensAntes);

            int alvo = alvos.get(caso);
            Balancing.ResultadoBalanceamento resultado = Balancing.balancearPorTemperatura(atual, caso, alpha, alvo);
            atual = resultado.exemplos();

            Map<String, Integer> contagensDepois = Balancing.contarPorFonte(atual, caso);
            Map<String, Double> distDepois = Diversity.distribuicaoDe(contagensDepois);

            relatorioPorCaso.put(caso, new RelatorioPorCaso(
                    contagensAntes, distAntes, Diversity.entropiaShannon(distAntes), Diversity.numeroEfetivoFontes(distAntes),
                    contagensDepois, distDepois, Diversity.entropiaShannon(distDepois), Diversity.numeroEfetivoFontes(distDepois),
                    resultado.alocacao()
            ));
        }

        return new ResultadoPipeline(
                exemplos.size(), dedup.mantidos().size(), dedup.removidos(),
                dedup.resultadosPorCaso(), dedup.totalParesForcaBruta(), dedup.totalCandidatosLSH(),
                atual.size(), relatorioPorCaso, atual
        );
    }
}
