package com.finetuningapi;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.util.Map;

/**
 * Implementacao real do VertexAiJobClient: obtem token via
 * `gcloud auth print-access-token` (processo externo, porte de obterTokenAcesso
 * do JS, que usa execSync) e chama a API REST aiplatform.googleapis.com
 * (java.net.http.HttpClient, porte do fetch original). So esta classe toca
 * rede/processo externo -- todo o resto do projeto e logica pura, testavel
 * sem gcloud instalado.
 */
public final class VertexAiHttpClient implements VertexAiJobClient {

    private final String projeto;
    private final String regiao;
    private final HttpClient http = HttpClient.newHttpClient();

    public VertexAiHttpClient(String projeto, String regiao) {
        this.projeto = projeto;
        this.regiao = regiao;
    }

    String obterTokenAcesso() throws IOException, InterruptedException {
        Process processo = new ProcessBuilder("gcloud", "auth", "print-access-token")
                .redirectErrorStream(false)
                .start();
        String saida;
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(processo.getInputStream(), StandardCharsets.UTF_8))) {
            saida = reader.readLine();
        }
        int codigo = processo.waitFor();
        if (codigo != 0 || saida == null) {
            throw new IOException("gcloud auth print-access-token falhou com codigo " + codigo);
        }
        return saida.trim();
    }

    @Override
    public Map<String, Object> consultarJob(String nomeJob) throws IOException, InterruptedException {
        String token = obterTokenAcesso();
        String url = "https://" + regiao + "-aiplatform.googleapis.com/v1/" + nomeJob;
        HttpRequest request = HttpRequest.newBuilder(URI.create(url))
                .header("Authorization", "Bearer " + token)
                .GET()
                .build();
        HttpResponse<String> resposta = http.send(request, HttpResponse.BodyHandlers.ofString());
        if (resposta.statusCode() / 100 != 2) {
            throw new IOException("Falha ao consultar job: " + resposta.statusCode());
        }
        return JsonUtil.parseObjeto(resposta.body());
    }

    @Override
    public Map<String, Object> criarJob(Map<String, Object> corpo) throws IOException, InterruptedException {
        String token = obterTokenAcesso();
        String url = "https://" + regiao + "-aiplatform.googleapis.com/v1/projects/" + projeto
                + "/locations/" + regiao + "/tuningJobs";
        HttpRequest request = HttpRequest.newBuilder(URI.create(url))
                .header("Authorization", "Bearer " + token)
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(JsonUtil.toJson(corpo)))
                .build();
        HttpResponse<String> resposta = http.send(request, HttpResponse.BodyHandlers.ofString());
        if (resposta.statusCode() / 100 != 2) {
            throw new IOException("Falha ao criar job: " + resposta.statusCode());
        }
        return JsonUtil.parseObjeto(resposta.body());
    }
}
