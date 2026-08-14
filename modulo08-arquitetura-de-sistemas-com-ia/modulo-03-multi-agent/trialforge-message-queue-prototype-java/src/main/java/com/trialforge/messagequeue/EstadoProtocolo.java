package com.trialforge.messagequeue;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Porte de criarEstadoProtocolo/revisarProtocolo (JS/Python).
 *
 * <p>O Protocolo e um recurso VERSIONADO e mutavel, nao um valor congelado —
 * uma emenda pode chegar enquanto ICF/CSR ja estao trabalhando com uma copia
 * mais antiga (Modulo 3.4: "O Protocolo versao 1 nao e apagado — ele e
 * versionado, preservado como historico, e uma versao 2 e criada com o
 * criterio corrigido.").</p>
 *
 * <p>Thread-safe: e lido (versao/criterios) pelas threads do ICF e do CSR e
 * escrito pela thread que processa a emenda etica, que roda de forma
 * independente do fluxo principal.</p>
 */
public final class EstadoProtocolo {

    private final Object trava = new Object();
    private volatile int versao = 1;
    private volatile Map<String, Integer> criterios;
    private final List<RevisaoHistorico> historico = new ArrayList<>();

    public EstadoProtocolo(Map<String, Integer> criteriosIniciais) {
        this.criterios = Map.copyOf(criteriosIniciais);
    }

    public int getVersao() {
        return versao;
    }

    public Map<String, Integer> getCriterios() {
        return criterios;
    }

    public List<RevisaoHistorico> getHistorico() {
        synchronized (trava) {
            return List.copyOf(historico);
        }
    }

    public void revisar(Map<String, Integer> novosCriterios, String motivo) {
        synchronized (trava) {
            historico.add(new RevisaoHistorico(versao, criterios, motivo));
            versao = versao + 1;
            Map<String, Integer> mesclado = new LinkedHashMap<>(criterios);
            mesclado.putAll(novosCriterios);
            criterios = Map.copyOf(mesclado);
        }
    }
}
