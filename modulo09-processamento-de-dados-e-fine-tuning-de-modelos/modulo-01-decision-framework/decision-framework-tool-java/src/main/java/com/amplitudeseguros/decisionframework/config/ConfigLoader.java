package com.amplitudeseguros.decisionframework.config;

import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.IOException;
import java.io.InputStream;

/** Carrega amplitude-seguros-casos.json do classpath — equivalente a carregarConfiguracao() em JS. */
public final class ConfigLoader {

    private static final String RESOURCE_NAME = "amplitude-seguros-casos.json";

    private ConfigLoader() {
    }

    public static AmplitudeConfig carregarConfiguracao() {
        ObjectMapper mapper = new ObjectMapper()
                .configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
        try (InputStream in = ConfigLoader.class.getClassLoader().getResourceAsStream(RESOURCE_NAME)) {
            if (in == null) {
                throw new IllegalStateException("recurso nao encontrado no classpath: " + RESOURCE_NAME);
            }
            return mapper.readValue(in, AmplitudeConfig.class);
        } catch (IOException e) {
            throw new IllegalStateException("falha ao carregar " + RESOURCE_NAME, e);
        }
    }
}
