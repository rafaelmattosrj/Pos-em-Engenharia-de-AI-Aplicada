package com.lorapeft;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.google.gson.JsonSyntaxException;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Porte de rank-adapter-comparison-tool.js (Modulo 4.3, companion, demo parte
 * 2). Mesmo padrao de AdapterComparison: parsing puro e portavel, disparo de
 * `python3 -m mlx_lm generate` via subprocesso depende de MLX + adaptadores
 * locais.
 */
public final class RankAdapterComparison {

    private RankAdapterComparison() {
    }

    public static final String MODELO_BASE = "mlx-community/gemma-4-e2b-it-bf16";
    public static final int MAX_TOKENS = 80;

    public static final String PROMPT = "Extraia segurado, placa e valor do orçamento de oficina abaixo.\n\n"
            + "BOA VISTA REPAROS AUTOMOTIVOS CNPJ 21.098.765/0001-32 Rua dos Mecanicos 310 Segurado: "
            + "Ricardo Alves Monteiro Placa do veiculo: JBR-9021 Data do sinistro: 09/06/2026 Valor das "
            + "pecas: R$ 1.850,00 Valor da mao de obra: R$ 970,00 Valor total do orcamento: R$ 2.820,00";

    private static final Gson GSON = new Gson();
    private static final Pattern TOKENS_PATTERN = Pattern.compile("Generation: (\\d+) tokens");

    public static JsonObject gabarito() {
        JsonObject g = new JsonObject();
        g.addProperty("segurado", "Ricardo Alves Monteiro");
        g.addProperty("placa", "JBR-9021");
        g.addProperty("valor", 2820);
        return g;
    }

    /** rank -> caminho do diretorio de adaptador, mesma convencao de nomes do original. */
    public static Map<String, String> adapters(String baseDir) {
        Map<String, String> adapters = new LinkedHashMap<>();
        adapters.put("rank 4", baseDir + "/mlx-adapters-rank4");
        adapters.put("rank 8", baseDir + "/mlx-adapters");
        adapters.put("rank 16", baseDir + "/mlx-adapters-rank16");
        return adapters;
    }

    public static List<String> montarArgumentos(String adapterPath) {
        List<String> args = new ArrayList<>(List.of(
                "-m", "mlx_lm", "generate",
                "--model", MODELO_BASE,
                "--prompt", PROMPT,
                "--max-tokens", String.valueOf(MAX_TOKENS)));
        if (adapterPath != null) {
            args.add("--adapter-path");
            args.add(adapterPath);
        }
        return args;
    }

    public record ResultadoParse(String texto, Integer tokens, JsonObject json) {
    }

    public static ResultadoParse parsearSaida(String textoSaida) {
        String[] blocos = textoSaida.split("==========");
        if (blocos.length < 3) {
            throw new IllegalArgumentException("saída fora do formato esperado");
        }
        String texto = blocos[1].trim();
        Matcher m = TOKENS_PATTERN.matcher(textoSaida);
        Integer tokens = m.find() ? Integer.valueOf(m.group(1)) : 0;
        JsonObject json;
        try {
            json = GSON.fromJson(texto, JsonObject.class);
        } catch (JsonSyntaxException ignored) {
            json = null;
        }
        return new ResultadoParse(texto, tokens, json);
    }

    public static boolean baterComGabarito(JsonObject json) {
        if (json == null) {
            return false;
        }
        JsonObject gabarito = gabarito();
        for (String chave : gabarito.keySet()) {
            if (!Objects.equals(json.get(chave), gabarito.get(chave))) {
                return false;
            }
        }
        return true;
    }

    public static String rodarReal(String adapterPath) throws IOException, InterruptedException {
        List<String> comando = new ArrayList<>();
        comando.add("python3");
        comando.addAll(montarArgumentos(adapterPath));
        Process processo = new ProcessBuilder(comando).start();
        String stdout = new String(processo.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
        String stderr = new String(processo.getErrorStream().readAllBytes(), StandardCharsets.UTF_8);
        int status = processo.waitFor();
        if (status != 0) {
            throw new IllegalStateException("mlx_lm generate falhou (" + status + "):\n" + stderr);
        }
        return stdout + "\n" + stderr;
    }
}
