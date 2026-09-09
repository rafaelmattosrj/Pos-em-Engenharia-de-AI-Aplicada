package com.finetuningapi;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Serializador/parser JSON recursivo minimo (objetos, arrays, String, Number,
 * Boolean, null) -- evita puxar uma biblioteca externa (Jackson/Gson) pra uma
 * ferramenta CLI standalone, mantendo o padrão "Maven puro" da skill de porte.
 * Suficiente pra montar corpo de requisição e ler resposta aninhada da API
 * REST da Vertex AI (job.tuningDataStats.supervisedTuningDataStats, etc.).
 */
final class JsonUtil {

    private JsonUtil() {
    }

    @SuppressWarnings("unchecked")
    static String toJson(Object valor) {
        if (valor == null) return "null";
        if (valor instanceof String s) return "\"" + escape(s) + "\"";
        if (valor instanceof Number || valor instanceof Boolean) return String.valueOf(valor);
        if (valor instanceof Map<?, ?> map) {
            StringBuilder sb = new StringBuilder("{");
            boolean primeiro = true;
            for (Map.Entry<?, ?> entry : map.entrySet()) {
                if (!primeiro) sb.append(',');
                primeiro = false;
                sb.append('"').append(escape(String.valueOf(entry.getKey()))).append("\":").append(toJson(entry.getValue()));
            }
            return sb.append('}').toString();
        }
        if (valor instanceof List<?> list) {
            StringBuilder sb = new StringBuilder("[");
            boolean primeiro = true;
            for (Object item : list) {
                if (!primeiro) sb.append(',');
                primeiro = false;
                sb.append(toJson(item));
            }
            return sb.append(']').toString();
        }
        throw new IllegalArgumentException("tipo nao suportado: " + valor.getClass());
    }

    private static String escape(String s) {
        return s.replace("\\", "\\\\").replace("\"", "\\\"").replace("\n", "\\n");
    }

    static Object parse(String json) {
        Parser parser = new Parser(json);
        Object resultado = parser.parseValor();
        parser.pularEspacos();
        return resultado;
    }

    @SuppressWarnings("unchecked")
    static Map<String, Object> parseObjeto(String json) {
        return (Map<String, Object>) parse(json);
    }

    private static final class Parser {
        private final String s;
        private int i;

        Parser(String s) {
            this.s = s;
            this.i = 0;
        }

        void pularEspacos() {
            while (i < s.length() && Character.isWhitespace(s.charAt(i))) i++;
        }

        Object parseValor() {
            pularEspacos();
            char c = s.charAt(i);
            if (c == '{') return parseObjetoInterno();
            if (c == '[') return parseArrayInterno();
            if (c == '"') return parseStringInterno();
            if (s.startsWith("true", i)) {
                i += 4;
                return Boolean.TRUE;
            }
            if (s.startsWith("false", i)) {
                i += 5;
                return Boolean.FALSE;
            }
            if (s.startsWith("null", i)) {
                i += 4;
                return null;
            }
            return parseNumeroInterno();
        }

        Map<String, Object> parseObjetoInterno() {
            Map<String, Object> resultado = new LinkedHashMap<>();
            i++; // {
            pularEspacos();
            if (s.charAt(i) == '}') {
                i++;
                return resultado;
            }
            while (true) {
                pularEspacos();
                String chave = parseStringInterno();
                pularEspacos();
                i++; // :
                Object valor = parseValor();
                resultado.put(chave, valor);
                pularEspacos();
                if (s.charAt(i) == ',') {
                    i++;
                    continue;
                }
                i++; // }
                break;
            }
            return resultado;
        }

        List<Object> parseArrayInterno() {
            List<Object> resultado = new ArrayList<>();
            i++; // [
            pularEspacos();
            if (s.charAt(i) == ']') {
                i++;
                return resultado;
            }
            while (true) {
                Object valor = parseValor();
                resultado.add(valor);
                pularEspacos();
                if (s.charAt(i) == ',') {
                    i++;
                    continue;
                }
                i++; // ]
                break;
            }
            return resultado;
        }

        String parseStringInterno() {
            StringBuilder sb = new StringBuilder();
            i++; // "
            while (s.charAt(i) != '"') {
                char c = s.charAt(i);
                if (c == '\\' && i + 1 < s.length()) {
                    i++;
                    char esc = s.charAt(i);
                    sb.append(switch (esc) {
                        case 'n' -> '\n';
                        case 't' -> '\t';
                        case 'r' -> '\r';
                        default -> esc;
                    });
                } else {
                    sb.append(c);
                }
                i++;
            }
            i++; // "
            return sb.toString();
        }

        Number parseNumeroInterno() {
            int inicio = i;
            while (i < s.length() && (Character.isDigit(s.charAt(i)) || s.charAt(i) == '-' || s.charAt(i) == '+'
                    || s.charAt(i) == '.' || s.charAt(i) == 'e' || s.charAt(i) == 'E')) {
                i++;
            }
            String bruto = s.substring(inicio, i);
            if (bruto.contains(".") || bruto.contains("e") || bruto.contains("E")) return Double.parseDouble(bruto);
            return Long.parseLong(bruto);
        }
    }
}
