package com.datasetprep.extraction;

import java.io.IOException;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Porte de extraction-to-jsonl-tool.js: pipeline de OCR (binario tesseract
 * real) + parsing tolerante a variacao de rotulo (regex) + validacao de
 * schema, transformando documento bruto num exemplo estruturado de treino.
 *
 * A chamada ao binario tesseract (ocrTexto/ocrConfiancaMedia) exige o
 * executavel instalado com o pacote de idioma "por" - nao roda em CI/teste
 * automatizado sem essa dependencia externa. As funcoes de parsing e
 * validacao de schema sao puras e testadas isoladamente (mesmo padrao do
 * original, que ja separa OCR de parsing).
 */
public final class OcrExtraction {

    private OcrExtraction() {
    }

    /** Roda o Tesseract em modo texto puro contra a imagem, em portugues. Requer o binario `tesseract` no PATH. */
    public static String ocrTexto(Path caminhoImagem) throws IOException, InterruptedException {
        Process p = new ProcessBuilder("tesseract", caminhoImagem.toString(), "stdout", "-l", "por")
                .redirectErrorStream(false)
                .start();
        String saida = new String(p.getInputStream().readAllBytes());
        p.waitFor();
        return saida.trim();
    }

    /** Roda o Tesseract em modo TSV para extrair a confianca media (0.0-1.0) das palavras reconhecidas. */
    public static double ocrConfiancaMedia(Path caminhoImagem) throws IOException, InterruptedException {
        Process p = new ProcessBuilder("tesseract", caminhoImagem.toString(), "stdout", "-l", "por", "tsv")
                .redirectErrorStream(false)
                .start();
        String tsv = new String(p.getInputStream().readAllBytes());
        p.waitFor();

        String[] linhas = tsv.trim().split("\n");
        double soma = 0;
        int total = 0;
        for (int i = 1; i < linhas.length; i++) { // pula cabecalho
            String[] campos = linhas[i].split("\t");
            if (campos.length >= 12) {
                double conf = Double.parseDouble(campos[10]);
                if (conf >= 0) {
                    soma += conf / 100;
                    total++;
                }
            }
        }
        return total == 0 ? 0 : soma / total;
    }

    /** Converte valor no formato brasileiro ("3.210,50") para numero (3210.50). */
    public static double parsearValorBRL(String texto) {
        String limpo = texto.replaceAll("[^\\d.,]", "").replace(".", "").replace(",", ".");
        return Double.parseDouble(limpo);
    }

    private static String extrairCampo(String texto, List<Pattern> padroes) {
        for (Pattern padrao : padroes) {
            Matcher m = padrao.matcher(texto);
            if (m.find()) return m.group(1).trim();
        }
        return null;
    }

    public record CamposAuto(String segurado, String placa, Double valor) {
    }

    public static CamposAuto parsearOrcamentoAuto(String textoOcr) {
        String segurado = extrairCampo(textoOcr, List.of(
                Pattern.compile("(?:nome do )?segurado:?\\s*(.+)", Pattern.CASE_INSENSITIVE)));
        String placa = extrairCampo(textoOcr, List.of(
                Pattern.compile("placa(?:\\s+do\\s+ve[ií]culo)?:?\\s*([A-Z0-9\\-]+)", Pattern.CASE_INSENSITIVE)));
        String valorTexto = extrairCampo(textoOcr, List.of(
                Pattern.compile("(?:valor total do reparo|total):?\\s*r\\$?\\s*([\\d.,]+)", Pattern.CASE_INSENSITIVE)));
        return new CamposAuto(segurado, placa, valorTexto != null ? parsearValorBRL(valorTexto) : null);
    }

    public record CamposSaude(String beneficiario, String procedimento, Double valor) {
    }

    public static CamposSaude parsearReciboSaude(String textoOcr) {
        String beneficiario = extrairCampo(textoOcr, List.of(
                Pattern.compile("(?:paciente/)?benefici[aá]rio:?\\s*(.+)", Pattern.CASE_INSENSITIVE)));
        String procedimento = extrairCampo(textoOcr, List.of(
                Pattern.compile("procedimento(?:\\s+realizado)?:?\\s*(.+)", Pattern.CASE_INSENSITIVE)));
        String valorTexto = extrairCampo(textoOcr, List.of(
                Pattern.compile("valor(?:\\s+cobrado)?:?\\s*r\\$?\\s*([\\d.,]+)", Pattern.CASE_INSENSITIVE)));
        return new CamposSaude(beneficiario, procedimento, valorTexto != null ? parsearValorBRL(valorTexto) : null);
    }

    private static final Map<String, List<String>> CAMPOS_OBRIGATORIOS = Map.of(
            "amplitude-auto", List.of("segurado", "placa", "valor"),
            "amplitude-saude-empresarial", List.of("beneficiario", "procedimento", "valor")
    );

    public record Validacao(boolean valido, List<String> erros) {
    }

    /** Valida se um exemplo estruturado (mapa campo->valor) tem todos os campos obrigatorios do seu caso, sem nulo nem valor invalido. */
    public static Validacao validarExemplo(String caso, Map<String, Object> saida) {
        List<String> erros = new ArrayList<>();
        for (String campo : CAMPOS_OBRIGATORIOS.get(caso)) {
            Object v = saida.get(campo);
            if (v == null || "".equals(v)) {
                erros.add("campo obrigatorio ausente: " + campo);
            }
        }
        Object valor = saida.get("valor");
        if (valor instanceof Number n) {
            double d = n.doubleValue();
            if (Double.isNaN(d) || d <= 0) erros.add("valor invalido: " + d);
        }
        return new Validacao(erros.isEmpty(), erros);
    }

    public static Map<String, Object> camposAutoParaMapa(CamposAuto c) {
        Map<String, Object> m = new HashMap<>();
        m.put("segurado", c.segurado());
        m.put("placa", c.placa());
        m.put("valor", c.valor());
        return m;
    }

    public static Map<String, Object> camposSaudeParaMapa(CamposSaude c) {
        Map<String, Object> m = new HashMap<>();
        m.put("beneficiario", c.beneficiario());
        m.put("procedimento", c.procedimento());
        m.put("valor", c.valor());
        return m;
    }
}
