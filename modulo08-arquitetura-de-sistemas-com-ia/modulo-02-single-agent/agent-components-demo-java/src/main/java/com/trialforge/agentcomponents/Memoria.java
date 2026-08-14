package com.trialforge.agentcomponents;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 1. MEMORIA — curto prazo (efemera) vs longo prazo (persistida).
 *
 * Porte 1:1 de agent-components-demo.js / agent_components_demo.py (secao 1).
 */
public final class Memoria {

    private Memoria() {
    }

    /** Uma mensagem de chat simples {role, content} — equivalente ao objeto literal do original. */
    public record Mensagem(String role, String content) {
    }

    /**
     * Vive so durante esta chamada: some quando o metodo termina, nunca e reaproveitada
     * pelo proximo protocolo que chegar.
     */
    public static List<Mensagem> memoriaCurtoPrazo(String protocolo) {
        return List.of(
                new Mensagem("system", "Você redige seções de ICF para estudos clínicos."),
                new Mensagem("user", "Protocolo: " + protocolo)
        );
    }

    /** Snapshot imutavel do historico acumulado de um usuario. */
    public record Historico(int interacoes, List<String> preferencias) {
    }

    /**
     * Simula um "banco" simples pra mostrar o principio de memoria persistida entre
     * chamadas — em producao seria um banco de dados de verdade.
     *
     * Adaptacao: no original (JS/Python) o "banco" e um Map/dict em escopo de modulo,
     * compartilhado por todas as chamadas do processo, e os testes chamam .clear() antes
     * de rodar. Aqui isso vira um objeto instanciavel: cada MemoriaLongoPrazo tem seu
     * proprio mapa, entao "resetar entre testes" e so criar uma instancia nova — mesmo
     * comportamento observado, sem depender de estado global mutavel estatico.
     */
    public static class MemoriaLongoPrazo {
        private final Map<String, Historico> bancoDeUsuarios = new ConcurrentHashMap<>();

        public Historico registrar(String usuarioId, String preferenciaNova) {
            Historico atual = bancoDeUsuarios.getOrDefault(usuarioId, new Historico(0, List.of()));
            int novasInteracoes = atual.interacoes() + 1;
            List<String> novasPreferencias = atual.preferencias();
            if (preferenciaNova != null) {
                List<String> copia = new ArrayList<>(atual.preferencias());
                copia.add(preferenciaNova);
                novasPreferencias = List.copyOf(copia);
            }
            Historico novo = new Historico(novasInteracoes, novasPreferencias);
            bancoDeUsuarios.put(usuarioId, novo);
            return novo;
        }
    }
}
