package com.amplitudeseguros.decisionframework.config;

import java.util.List;

/** Raiz do amplitude-seguros-casos.json: limiar de aprovação, matriz AHP e os casos de negócio. */
public record AmplitudeConfig(double limiarVerde, AhpConfig ahp, List<Caso> casos) {

    public Caso caso(String id) {
        return casos.stream()
                .filter(c -> c.id().equals(id))
                .findFirst()
                .orElseThrow(() -> new IllegalArgumentException("caso nao encontrado: " + id));
    }
}
