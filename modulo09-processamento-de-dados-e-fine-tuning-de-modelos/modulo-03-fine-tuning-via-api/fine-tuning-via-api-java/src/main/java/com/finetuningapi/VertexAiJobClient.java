package com.finetuningapi;

import java.io.IOException;
import java.util.Map;

/**
 * Contrato de acesso a job de fine-tuning da Vertex AI. Injetavel pra teste
 * (mesmo principio de consultarFn/criarJobFn injetaveis nos originais JS) --
 * a implementacao real (VertexAiHttpClient) faz chamada de rede de verdade,
 * a de teste devolve estado fixo sem tocar rede.
 */
public interface VertexAiJobClient {
    Map<String, Object> consultarJob(String nomeJob) throws IOException, InterruptedException;

    Map<String, Object> criarJob(Map<String, Object> corpo) throws IOException, InterruptedException;
}
