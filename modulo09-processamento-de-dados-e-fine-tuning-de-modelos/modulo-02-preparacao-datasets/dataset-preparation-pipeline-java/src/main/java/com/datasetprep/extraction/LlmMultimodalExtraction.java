package com.datasetprep.extraction;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.file.Files;
import java.nio.file.Path;
import java.text.Normalizer;
import java.time.Duration;
import java.util.Base64;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.regex.Pattern;

/**
 * Porte de extracao-llm-multimodal-tool.js: extracao via LLM multimodal
 * (Gemini/Vertex AI), em oposicao ao pipeline de OCR classico de
 * OcrExtraction. Manda a imagem direto pro Gemini via Vertex AI, pedindo
 * JSON estruturado sem nenhum passo de regex intermediario.
 *
 * Requer: `gcloud auth application-default login` com acesso ao projeto GCP
 * (obterTokenAcesso invoca o binario `gcloud`), e rede - nao roda em
 * CI/teste automatizado sem essas dependencias externas. compararComEsperado
 * e normalizar sao puras e testadas isoladamente, igual ao original.
 */
public final class LlmMultimodalExtraction {

    private LlmMultimodalExtraction() {
    }

    private static final String PROJETO = "amplitude-seguros-demo";
    private static final String REGIAO = "us-central1";
    private static final String MODELO = "gemini-2.5-flash";

    /** Obtem o token de acesso via `gcloud auth print-access-token`. Requer gcloud autenticado nesta maquina. */
    public static String obterTokenAcesso() throws IOException, InterruptedException {
        Process p = new ProcessBuilder("gcloud", "auth", "print-access-token").start();
        String saida = new String(p.getInputStream().readAllBytes()).trim();
        int status = p.waitFor();
        if (status != 0 || saida.isEmpty()) {
            throw new IllegalStateException("Nao consegui obter um token de acesso via gcloud. "
                    + "Rode \"gcloud auth login\" nesta maquina antes de rodar a demo ao vivo.");
        }
        return saida;
    }

    private static final Pattern MARCAS_DIACRITICAS = Pattern.compile("\\p{InCombiningDiacriticalMarks}+");

    public static String normalizar(Object texto) {
        String s = String.valueOf(texto);
        String semAcento = MARCAS_DIACRITICAS.matcher(Normalizer.normalize(s, Normalizer.Form.NFD)).replaceAll("");
        return semAcento.toLowerCase().trim().replaceAll("\\s+", " ");
    }

    public record Resultado(Map<String, Object> extraido, long latenciaMs, Long tokensEntrada, Long tokensSaida) {
    }

    /** Chama o Gemini multimodal via Vertex AI com a imagem + instrucao, pedindo JSON estruturado. */
    public static Resultado extrairViaLlm(Path caminhoImagem, String instrucao, List<String> campos) throws IOException, InterruptedException {
        String imagemBase64 = Base64.getEncoder().encodeToString(Files.readAllBytes(caminhoImagem));
        String listaCampos = String.join(", ", campos);

        String prompt = instrucao + " Devolva SOMENTE um JSON valido, sem markdown "
                + "e sem texto extra, com exatamente estas chaves: " + listaCampos + ". O campo \"valor\" deve "
                + "ser um numero (nao string, sem simbolo de moeda).";

        String token = obterTokenAcesso();
        String url = "https://" + REGIAO + "-aiplatform.googleapis.com/v1/projects/" + PROJETO
                + "/locations/" + REGIAO + "/publishers/google/models/" + MODELO + ":generateContent";

        String corpo = "{"
                + "\"contents\":[{\"role\":\"user\",\"parts\":["
                + "{\"text\":" + jsonString(prompt) + "},"
                + "{\"inline_data\":{\"mime_type\":\"image/png\",\"data\":\"" + imagemBase64 + "\"}}"
                + "]}],"
                + "\"generationConfig\":{\"temperature\":0,\"responseMimeType\":\"application/json\"}"
                + "}";

        HttpClient client = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(30)).build();
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(url))
                .header("Authorization", "Bearer " + token)
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(corpo))
                .build();

        long inicio = System.currentTimeMillis();
        HttpResponse<String> resposta = client.send(request, HttpResponse.BodyHandlers.ofString());
        long latenciaMs = System.currentTimeMillis() - inicio;

        if (resposta.statusCode() < 200 || resposta.statusCode() >= 300) {
            throw new IllegalStateException("Vertex AI retornou " + resposta.statusCode() + ": " + resposta.body());
        }

        // Parsing minimo do envelope de resposta do Vertex AI (sem lib externa de JSON).
        String corpoResposta = resposta.body();
        String textoResposta = extrairTextoDaResposta(corpoResposta);
        if (textoResposta == null) {
            throw new IllegalStateException("resposta sem texto utilizavel: " + corpoResposta);
        }

        Map<String, Object> extraido = parsearJsonPlano(textoResposta);
        Long tokensEntrada = extrairLong(corpoResposta, "promptTokenCount");
        Long tokensSaida = extrairLong(corpoResposta, "candidatesTokenCount");

        return new Resultado(extraido, latenciaMs, tokensEntrada, tokensSaida);
    }

    /** Compara o extraido com o gabarito esperado, campo a campo, com tolerancia a acento/caixa para strings. */
    public static Map<String, Object> compararComEsperado(Map<String, Object> extraido, Map<String, Object> esperado) {
        Map<String, Object> porCampo = new LinkedHashMap<>();
        int acertos = 0;
        for (Map.Entry<String, Object> e : esperado.entrySet()) {
            String campo = e.getKey();
            Object valorEsperado = e.getValue();
            Object valorExtraido = extraido.get(campo);
            boolean bate;
            if (valorEsperado instanceof Number esperadoNum) {
                bate = valorExtraido != null && Double.parseDouble(valorExtraido.toString()) == esperadoNum.doubleValue();
            } else {
                bate = normalizar(valorExtraido == null ? "" : valorExtraido).equals(normalizar(valorEsperado));
            }
            Map<String, Object> detalhe = new LinkedHashMap<>();
            detalhe.put("esperado", valorEsperado);
            detalhe.put("extraido", valorExtraido);
            detalhe.put("bate", bate);
            porCampo.put(campo, detalhe);
            if (bate) acertos++;
        }
        Map<String, Object> resultado = new LinkedHashMap<>();
        resultado.put("porCampo", porCampo);
        resultado.put("acertos", acertos);
        resultado.put("total", esperado.size());
        return resultado;
    }

    // --- utilitarios minimos de JSON (sem dependencia externa) ---

    private static String jsonString(String s) {
        return "\"" + s.replace("\\", "\\\\").replace("\"", "\\\"") + "\"";
    }

    private static String extrairTextoDaResposta(String corpoJson) {
        // Procura "text": "..." dentro de candidates[0].content.parts[0] (JSON simples, sem aninhamento de aspas no texto).
        Pattern p = Pattern.compile("\"text\"\\s*:\\s*\"((?:[^\"\\\\]|\\\\.)*)\"");
        var m = p.matcher(corpoJson);
        if (!m.find()) return null;
        return m.group(1).replace("\\\"", "\"").replace("\\\\", "\\").replace("\\n", "\n");
    }

    private static Long extrairLong(String corpoJson, String chave) {
        Pattern p = Pattern.compile("\"" + chave + "\"\\s*:\\s*(\\d+)");
        var m = p.matcher(corpoJson);
        return m.find() ? Long.parseLong(m.group(1)) : null;
    }

    /** Parser de JSON plano (objeto de 1 nivel, valores string/numero) - suficiente para o retorno esperado do prompt. */
    private static Map<String, Object> parsearJsonPlano(String json) {
        Map<String, Object> resultado = new LinkedHashMap<>();
        Pattern par = Pattern.compile("\"([^\"]+)\"\\s*:\\s*(?:\"((?:[^\"\\\\]|\\\\.)*)\"|(-?\\d+(?:\\.\\d+)?))");
        var m = par.matcher(json);
        while (m.find()) {
            String chave = m.group(1);
            if (m.group(2) != null) {
                resultado.put(chave, m.group(2).replace("\\\"", "\"").replace("\\\\", "\\"));
            } else {
                resultado.put(chave, Double.parseDouble(m.group(3)));
            }
        }
        return resultado;
    }
}
