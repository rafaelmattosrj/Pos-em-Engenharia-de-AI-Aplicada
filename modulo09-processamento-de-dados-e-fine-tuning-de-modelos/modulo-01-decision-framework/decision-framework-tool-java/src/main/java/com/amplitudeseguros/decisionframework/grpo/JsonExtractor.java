package com.amplitudeseguros.decisionframework.grpo;

import com.fasterxml.jackson.databind.ObjectMapper;

import java.util.Map;
import java.util.Optional;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/** Extrai o primeiro objeto JSON solto num texto -- equivalente a extrairJson() em grpo-verifiable-reward-demo.js. */
public final class JsonExtractor {

    private static final Pattern OBJETO_JSON = Pattern.compile("\\{[\\s\\S]*\\}");
    private static final ObjectMapper MAPPER = new ObjectMapper();

    private JsonExtractor() {
    }

    public static Optional<Map<String, Object>> extrairJson(String texto) {
        Matcher matcher = OBJETO_JSON.matcher(texto);
        if (!matcher.find()) {
            return Optional.empty();
        }
        try {
            return Optional.of(MAPPER.readValue(matcher.group(), Map.class));
        } catch (Exception e) {
            return Optional.empty();
        }
    }
}
