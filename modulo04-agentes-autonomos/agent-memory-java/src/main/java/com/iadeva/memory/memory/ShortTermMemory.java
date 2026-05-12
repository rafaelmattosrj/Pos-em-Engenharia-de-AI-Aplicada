package com.iadeva.memory.memory;

// Memória curta — equivalente ao short_term_memory.py — estado volátil da execução atual

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.context.annotation.Scope;
import org.springframework.stereotype.Component;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

/**
 * Memória de curto prazo: armazena estado temporário da execução atual.
 * Escopo prototype — cada execução do agente obtém uma instância nova,
 * garantindo isolamento total entre chamadas. Descartada ao fim do ciclo.
 */
@Component
@Scope(ConfigurableBeanFactory.SCOPE_PROTOTYPE)
public class ShortTermMemory {

    // Mapa volátil — vive apenas durante a execução atual
    private final Map<String, Object> store = new HashMap<>();

    /**
     * Armazena um valor associado à chave.
     */
    public void put(String key, Object value) {
        store.put(key, value);
    }

    /**
     * Recupera um valor pela chave. Retorna null se não encontrado.
     */
    public Object get(String key) {
        return store.get(key);
    }

    /**
     * Retorna visão somente-leitura de todo o estado atual.
     */
    public Map<String, Object> getAll() {
        return Collections.unmodifiableMap(store);
    }

    /**
     * Limpa todo o estado — chamado ao fim da execução para liberar memória.
     */
    public void clear() {
        store.clear();
    }
}
