package com.lorapeft;

import com.google.gson.Gson;
import com.google.gson.JsonObject;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Porte de lora-managed-api-preview-tool.js (Modulo 4.2, companion). Monta a
 * MESMA requisicao HTTP real (URL, headers, corpo) que a API de fine-tuning
 * da Together AI espera para um job LoRA (config real reaproveitada do treino
 * local do M4.2: rank 8, scale 20.0, dropout 0.0). So envia de verdade se
 * TOGETHER_API_KEY estiver no ambiente; sem a chave, so monta e imprime a
 * requisicao (modo preview), igual ao original.
 *
 * <p>Ressalva de honestidade herdada do original: "scale" (MLX-LM) e
 * "lora_alpha" (Together AI) nao sao garantidamente definidos de forma
 * identica entre frameworks -- o mesmo valor numerico e usado aqui como ponte
 * ilustrativa, nao equivalencia matematica comprovada.</p>
 */
public final class LoraManagedApiPreview {

    private LoraManagedApiPreview() {
    }

    public static final String TOGETHER_ENDPOINT = "https://api.together.ai/v1/fine-tunes";
    public static final String MODELO_PADRAO = "meta-llama/Meta-Llama-3.1-8B-Instruct-Reference";

    public record ConfigLora(int rank, double dropout, double scale) {
    }

    public static final ConfigLora CONFIG_TREINADA_LOCAL = new ConfigLora(8, 0.0, 20.0);

    public record Requisicao(String url, String method, Map<String, String> headers, JsonObject body) {
    }

    public static Requisicao montarRequisicaoLoraGerenciada(String apiKey, String trainingFileId, String modelo, ConfigLora loraConfig) {
        if (trainingFileId == null || trainingFileId.isBlank()) {
            throw new IllegalArgumentException("trainingFileId é obrigatório (id do arquivo já enviado à API)");
        }

        JsonObject trainingType = new JsonObject();
        trainingType.addProperty("type", "Lora");
        trainingType.addProperty("lora_r", loraConfig.rank());
        trainingType.addProperty("lora_alpha", loraConfig.scale());
        trainingType.addProperty("lora_dropout", loraConfig.dropout());
        trainingType.addProperty("lora_trainable_modules", "all-linear");

        JsonObject body = new JsonObject();
        body.addProperty("model", modelo);
        body.addProperty("training_file", trainingFileId);
        body.add("training_type", trainingType);

        Map<String, String> headers = new LinkedHashMap<>();
        headers.put("Authorization", "Bearer " + (apiKey != null ? apiKey : "<TOGETHER_API_KEY>"));
        headers.put("Content-Type", "application/json");

        return new Requisicao(TOGETHER_ENDPOINT, "POST", headers, body);
    }

    public static Requisicao montarRequisicaoLoraGerenciada(String apiKey, String trainingFileId) {
        return montarRequisicaoLoraGerenciada(apiKey, trainingFileId, MODELO_PADRAO, CONFIG_TREINADA_LOCAL);
    }

    public sealed interface Resultado permits Preview, Real {
    }

    public record Preview(Requisicao requisicao) implements Resultado {
    }

    public record Real(Requisicao requisicao, String respostaJson) implements Resultado {
    }

    public static Resultado enviarOuPrever(String trainingFileId) throws IOException, InterruptedException {
        String apiKey = System.getenv("TOGETHER_API_KEY");
        Requisicao requisicao = montarRequisicaoLoraGerenciada(apiKey, trainingFileId);

        if (apiKey == null || apiKey.isBlank()) {
            return new Preview(requisicao);
        }

        HttpClient client = HttpClient.newHttpClient();
        HttpRequest.Builder builder = HttpRequest.newBuilder(URI.create(requisicao.url()))
                .POST(HttpRequest.BodyPublishers.ofString(new Gson().toJson(requisicao.body())));
        requisicao.headers().forEach(builder::header);

        HttpResponse<String> resposta = client.send(builder.build(), HttpResponse.BodyHandlers.ofString());
        if (resposta.statusCode() < 200 || resposta.statusCode() >= 300) {
            throw new IllegalStateException("Together AI recusou a requisição (" + resposta.statusCode() + "): " + resposta.body());
        }
        return new Real(requisicao, resposta.body());
    }
}
