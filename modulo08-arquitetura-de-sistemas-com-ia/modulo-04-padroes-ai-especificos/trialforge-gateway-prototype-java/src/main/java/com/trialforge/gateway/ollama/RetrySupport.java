package com.trialforge.gateway.ollama;

import com.trialforge.gateway.DemoLogger;

import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

/**
 * comTimeout + comRetry — mesmo padrao de retry com limite do Modulo 3.5,
 * aqui aplicado as duas chamadas ao Ollama (embedding e geracao) que antes
 * nao tinham nenhuma protecao contra travamento ou instabilidade momentanea
 * do modelo local. Adaptado pra chamada bloqueante (a versao JS usa
 * Promise.race, aqui roda a chamada numa thread separada e usa
 * Future.get(timeout) como temporizador — mesma ideia do
 * ThreadPoolExecutor(...).result(timeout=...) da versao Python). Limitacao
 * honesta identica a da versao Python: se a chamada travar de verdade, nao
 * da pra matar a thread a forca, ela so termina sozinha quando (ou se) o
 * Ollama eventualmente responder — o timeout aqui desiste de ESPERAR a
 * resposta, nao cancela a chamada em si.
 */
public final class RetrySupport {

    private RetrySupport() {
    }

    public static <T> T comRetry(Callable<T> chamada, int maxTentativas, long timeoutMs, String operacao,
            DemoLogger logger) throws Exception {
        Exception ultimoErro = null;
        for (int tentativa = 1; tentativa <= maxTentativas; tentativa++) {
            ExecutorService executor = Executors.newSingleThreadExecutor(r -> {
                Thread t = new Thread(r, "ollama-" + operacao);
                t.setDaemon(true);
                return t;
            });
            try {
                Future<T> futuro = executor.submit(chamada);
                return futuro.get(timeoutMs, TimeUnit.MILLISECONDS);
            } catch (TimeoutException e) {
                ultimoErro = new OllamaTimeoutException(operacao, timeoutMs);
            } catch (ExecutionException e) {
                ultimoErro = (e.getCause() instanceof Exception ex) ? ex : e;
            } finally {
                executor.shutdown();
            }
            if (logger != null) {
                logger.log(String.format("[Retry] %s falhou na tentativa %d/%d: %s",
                        operacao, tentativa, maxTentativas, ultimoErro.getMessage()));
            }
        }
        throw new Exception(String.format("%s falhou após %d tentativas: %s", operacao, maxTentativas,
                ultimoErro.getMessage()), ultimoErro);
    }
}
