package com.trialforge.reactagent;

import com.fasterxml.jackson.databind.node.ObjectNode;

import java.io.IOException;
import java.util.List;
import java.util.Locale;

/**
 * Decorador de retry com backoff (chamada ao modelo, nao a ferramenta): distingue
 * falha transitoria (rede, timeout — vale tentar de novo) de falha terminal (ex.:
 * modelo nao existe no Ollama — retry nao resolve, e erro de configuracao). Sem
 * isso, qualquer soneca de rede escalava pro Approval Gate sem necessidade.
 *
 * Porte 1:1 de chamarModeloComRetry / ehErroTransitorio em
 * react-agent-prototype.js / react_agent_prototype.py.
 */
public class RetryingChatClient implements ChatClient {

    private static final int TENTATIVAS_MAX_PADRAO = 3;

    private final ChatClient delegate;
    private final int tentativasMax;

    public RetryingChatClient(ChatClient delegate) {
        this(delegate, TENTATIVAS_MAX_PADRAO);
    }

    public RetryingChatClient(ChatClient delegate, int tentativasMax) {
        this.delegate = delegate;
        this.tentativasMax = tentativasMax;
    }

    @Override
    public ModeloResposta chat(List<ObjectNode> historico, List<ObjectNode> tools)
            throws IOException, InterruptedException {
        IOException ultimoErro = null;

        for (int tentativa = 1; tentativa <= tentativasMax; tentativa++) {
            try {
                return delegate.chat(historico, tools);
            } catch (IOException erro) {
                ultimoErro = erro;
                if (!ehErroTransitorio(erro) || tentativa == tentativasMax) {
                    throw erro; // erro terminal, ou ja esgotou as tentativas: sobe pro catch-all
                }
                long esperaMs = 500L * (1L << (tentativa - 1)); // 500ms, 1s, 2s...
                System.out.println("[Retry] Falha transitória na chamada ao modelo (tentativa " + tentativa + "/"
                        + tentativasMax + "): " + erro.getMessage() + ". Tentando de novo em " + esperaMs + "ms.");
                Thread.sleep(esperaMs);
            }
        }

        throw ultimoErro;
    }

    /**
     * Adaptacao: no original, {@code erro.status_code === 404} e o unico caso
     * explicitamente terminal; qualquer outro codigo HTTP cairia no teste generico
     * de mensagem (que nunca bate pra um erro puramente de status). Aqui isso vira
     * "qualquer {@link ModeloHttpException} e terminal" — mesmo resultado observavel
     * (nenhum status HTTP e tratado como transitorio), mais direto de expressar em Java.
     */
    static boolean ehErroTransitorio(Throwable erro) {
        if (erro instanceof ModeloHttpException) {
            return false; // erro de configuração/servidor: retry não resolve
        }
        String mensagem = String.valueOf(erro.getMessage()).toLowerCase(Locale.ROOT);
        return erro instanceof java.net.ConnectException
                || erro instanceof java.net.http.HttpTimeoutException
                || erro instanceof java.net.SocketTimeoutException
                || mensagem.contains("timeout")
                || mensagem.contains("connection refused")
                || mensagem.contains("connection reset")
                || mensagem.contains("fetch failed")
                || mensagem.contains("econnrefused");
    }
}
