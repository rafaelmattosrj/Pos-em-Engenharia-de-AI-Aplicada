package com.lorapeft;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.google.gson.JsonSyntaxException;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Porte de adapter-comparison-tool.js (Modulo 4.2, companion). O parsing e a
 * montagem de argumentos sao logica pura e 100% portavel; ja rodarGenerateReal
 * dispara um processo `python3 -m mlx_lm generate`, que so funciona com MLX +
 * modelo/adaptador presentes localmente (mesma dependencia de hardware do
 * script original -- nao ha equivalente de biblioteca Java, so o disparo de
 * processo via ProcessBuilder e portavel, nao a inferencia MLX em si).
 */
public final class AdapterComparison {

    private AdapterComparison() {
    }

    public static final String MODELO_BASE = "mlx-community/gemma-4-e2b-it-bf16";
    public static final int MAX_TOKENS = 80;
    public static final int INDICE_EXEMPLO = 8;

    private static final Pattern TOKENS_PATTERN = Pattern.compile("Generation: (\\d+) tokens");
    private static final Gson GSON = new Gson();

    public record ExemploTeste(String prompt, JsonObject gabarito) {
    }

    public record ResultadoParse(String textoGerado, Integer tokensGerados, JsonObject json, boolean bateuNoLimiteDeTokens) {
    }

    public static ExemploTeste carregarExemploTeste(Path testJsonl, int indice) throws IOException {
        List<String> linhas = new ArrayList<>();
        for (String linha : Files.readAllLines(testJsonl, StandardCharsets.UTF_8)) {
            if (!linha.isBlank()) {
                linhas.add(linha);
            }
        }
        if (indice < 0 || indice >= linhas.size()) {
            throw new IllegalArgumentException("índice " + indice + " fora do intervalo (0-" + (linhas.size() - 1) + ")");
        }
        JsonObject exemplo = GSON.fromJson(linhas.get(indice), JsonObject.class);
        var messages = exemplo.getAsJsonArray("messages");
        String prompt = messages.get(0).getAsJsonObject().get("content").getAsString();
        JsonObject gabarito = GSON.fromJson(messages.get(1).getAsJsonObject().get("content").getAsString(), JsonObject.class);
        return new ExemploTeste(prompt, gabarito);
    }

    public static List<String> montarArgumentosGenerate(String prompt, String adapterPath, int maxTokens, String modelo) {
        List<String> args = new ArrayList<>(List.of(
                "-m", "mlx_lm", "generate",
                "--model", modelo,
                "--prompt", prompt,
                "--max-tokens", String.valueOf(maxTokens)));
        if (adapterPath != null) {
            args.add("--adapter-path");
            args.add(adapterPath);
        }
        return args;
    }

    public static ResultadoParse parsearSaidaGenerate(String textoSaida) {
        String[] blocos = textoSaida.split("==========");
        if (blocos.length < 3) {
            throw new IllegalArgumentException("saída não tem o formato esperado (dois separadores \"==========\")");
        }
        String textoGerado = blocos[1].trim();
        Matcher m = TOKENS_PATTERN.matcher(textoSaida);
        Integer tokensGerados = m.find() ? Integer.valueOf(m.group(1)) : null;
        JsonObject json;
        try {
            json = GSON.fromJson(textoGerado, JsonObject.class);
        } catch (JsonSyntaxException ignored) {
            json = null;
        }
        boolean bateuNoLimite = tokensGerados != null && tokensGerados >= MAX_TOKENS;
        return new ResultadoParse(textoGerado, tokensGerados, json, bateuNoLimite);
    }

    public static boolean compararComGabarito(JsonObject json, JsonObject gabarito) {
        if (json == null) {
            return false;
        }
        for (String chave : gabarito.keySet()) {
            if (!Objects.equals(json.get(chave), gabarito.get(chave))) {
                return false;
            }
        }
        return true;
    }

    public static String rodarGenerateReal(List<String> args) throws IOException, InterruptedException {
        List<String> comando = new ArrayList<>();
        comando.add("python3");
        comando.addAll(args);
        Process processo = new ProcessBuilder(comando).start();
        String stdout = new String(processo.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
        String stderr = new String(processo.getErrorStream().readAllBytes(), StandardCharsets.UTF_8);
        int status = processo.waitFor();
        if (status != 0) {
            throw new IllegalStateException("mlx_lm generate saiu com código " + status + ":\n" + stderr);
        }
        return stdout + "\n" + stderr;
    }
}
