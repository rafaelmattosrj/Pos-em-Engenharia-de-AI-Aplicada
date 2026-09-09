package com.finetuningapi;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Extra: dataset real alternativo (databricks-dolly-15k, CC-BY-SA-3.0) --
 * porte de dolly-dataset-real-starter.js. Conceitualmente equivalente ao
 * Modulo 2.2 (preparacao de dataset), so que contra dado real, nao
 * sintetico. Requer baixar o arquivo (13MB, ~15 mil linhas JSONL) antes de
 * rodar -- ver README deste projeto.
 */
public final class DollyDatasetStarter {

    public static final List<String> CATEGORIAS_COMPATIVEIS = List.of("information_extraction", "closed_qa", "summarization");
    public static final String CASO = "dolly-instruction-tuning";
    public static final int N_SHINGLE = 5;

    private DollyDatasetStarter() {
    }

    public record RegistroDolly(String instruction, String context, String response, String category) {
    }

    public static List<RegistroDolly> carregarDolly(Path caminhoJsonl) throws IOException {
        List<String> linhas = Files.readAllLines(caminhoJsonl, StandardCharsets.UTF_8);
        List<RegistroDolly> registros = new ArrayList<>();
        for (String linha : linhas) {
            if (linha.isBlank()) continue;
            Map<String, Object> obj = JsonUtil.parseObjeto(linha);
            registros.add(new RegistroDolly(
                    (String) obj.get("instruction"), (String) obj.get("context"),
                    (String) obj.get("response"), (String) obj.get("category")));
        }
        return registros;
    }

    public static List<DatasetScaling.ExemploDataset> paraSchemaCanonico(List<RegistroDolly> registros) {
        List<DatasetScaling.ExemploDataset> mapeados = new ArrayList<>();
        int indice = 0;
        for (RegistroDolly r : registros) {
            Map<String, Object> saida = new LinkedHashMap<>();
            saida.put("texto", r.response());
            mapeados.add(new DatasetScaling.ExemploDataset(
                    r.instruction(), r.context(), saida, CASO, r.category(),
                    "dolly-" + r.category() + "-" + indice));
            indice++;
        }
        return mapeados;
    }

    public static List<RegistroDolly> filtrarCompativeis(List<RegistroDolly> registros) {
        return registros.stream()
                .filter(r -> CATEGORIAS_COMPATIVEIS.contains(r.category())
                        && r.context() != null && !r.context().isBlank()
                        && r.response() != null && !r.response().isBlank())
                .toList();
    }

    private static String textoParaDedup(DatasetScaling.ExemploDataset e) {
        return e.instrucao() + "\n" + e.entrada();
    }

    public record ResultadoPreparo(int bruto, int compativeis, int mapeados, int itensRemovidos,
                                    int semDuplicatas, Map<String, Integer> contagem, Map<String, Integer> alocacao, int balanceado) {
    }

    /** Pipeline completo (dedup no dataset inteiro + balanceamento por temperatura), porte de prepararDatasetCompleto. */
    public static ResultadoPreparo prepararDatasetCompleto(Path caminhoJsonl, int alvoTotal) throws IOException {
        List<RegistroDolly> bruto = carregarDolly(caminhoJsonl);
        List<RegistroDolly> compativeis = filtrarCompativeis(bruto);
        List<DatasetScaling.ExemploDataset> mapeados = paraSchemaCanonico(compativeis);

        List<MinHashDedupBalancer.Exemplo> paraDedup = mapeados.stream()
                .map(e -> new MinHashDedupBalancer.Exemplo(e.id(), e.caso(), e.fonte(), textoParaDedup(e)))
                .toList();
        var dedup = MinHashDedupBalancer.encontrarQuaseDuplicatasGenerico(paraDedup, N_SHINGLE);
        java.util.Set<Integer> remover = new java.util.HashSet<>();
        for (var par : dedup.paresDuplicata()) remover.add(par.j());

        List<DatasetScaling.ExemploDataset> semDuplicatas = new ArrayList<>();
        for (int i = 0; i < mapeados.size(); i++) if (!remover.contains(i)) semDuplicatas.add(mapeados.get(i));

        Map<String, Integer> contagem = new LinkedHashMap<>();
        for (DatasetScaling.ExemploDataset e : semDuplicatas) contagem.merge(e.fonte(), 1, Integer::sum);
        Map<String, Integer> alocacao = MinHashDedupBalancer.alocarComCapacidade(contagem, MinHashDedupBalancer.ALPHA_TEMPERATURA, alvoTotal);

        Map<String, Integer> usados = new LinkedHashMap<>();
        int balanceado = 0;
        for (DatasetScaling.ExemploDataset e : semDuplicatas) {
            int usadoAtual = usados.getOrDefault(e.fonte(), 0);
            if (usadoAtual < alocacao.getOrDefault(e.fonte(), 0)) {
                usados.put(e.fonte(), usadoAtual + 1);
                balanceado++;
            }
        }

        return new ResultadoPreparo(bruto.size(), compativeis.size(), mapeados.size(), remover.size(),
                semDuplicatas.size(), contagem, alocacao, balanceado);
    }
}
