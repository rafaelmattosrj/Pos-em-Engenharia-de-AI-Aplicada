package com.trialforge.gateway;

/**
 * Log simples usado pra reproduzir a mesma saida de console dos originais —
 * quem chama ja formata a mensagem (String.format) antes de logar.
 */
@FunctionalInterface
public interface DemoLogger {
    void log(String mensagem);
}
