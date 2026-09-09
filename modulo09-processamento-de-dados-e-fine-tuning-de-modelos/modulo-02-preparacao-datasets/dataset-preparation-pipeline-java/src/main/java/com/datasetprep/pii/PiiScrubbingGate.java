package com.datasetprep.pii;

import java.util.ArrayList;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Porte de pii-scrubbing-gate-tool.js: gate de higienizacao de PII antes que
 * um documento entre no dataset de fine-tuning. Abordagem hibrida (regex +
 * heuristica), a mesma usada em pipelines reais de producao (Microsoft
 * Presidio combina regex + NER).
 *   - CPF: regex de formato + validacao real do digito verificador (Modulo 11).
 *   - Nome: ancora de rotulo (Segurado:/Beneficiario:) mais confiavel.
 */
public final class PiiScrubbingGate {

    private PiiScrubbingGate() {
    }

    private static final Pattern REGEX_CPF = Pattern.compile("\\b\\d{3}\\.?\\d{3}\\.?\\d{3}-?\\d{2}\\b");

    private static int calcularDigitoVerificador(String parcial) {
        int soma = 0;
        int peso = parcial.length() + 1;
        for (char c : parcial.toCharArray()) {
            soma += Character.getNumericValue(c) * peso;
            peso -= 1;
        }
        int resto = soma % 11;
        return resto < 2 ? 0 : 11 - resto;
    }

    public static boolean validarCPF(String cpfComOuSemMascara) {
        String digitos = cpfComOuSemMascara.replaceAll("\\D", "");
        if (digitos.length() != 11) return false;
        if (digitos.chars().distinct().count() == 1) return false; // todos os digitos iguais
        int d1 = calcularDigitoVerificador(digitos.substring(0, 9));
        int d2 = calcularDigitoVerificador(digitos.substring(0, 9) + d1);
        return digitos.charAt(9) == Character.forDigit(d1, 10) && digitos.charAt(10) == Character.forDigit(d2, 10);
    }

    private static final Pattern REGEX_NOME_ANCORADO = Pattern.compile(
            "(Segurado|Beneficiário|Beneficiario|Nome do segurado)[ \\t]*:[ \\t]*"
                    + "([A-ZÀ-Ú][\\wÀ-ú]*(?:[ \\t]+[A-ZÀ-Ú][\\wÀ-ú]*){1,4})"
    );

    public record NomeEncontrado(String nome, String confianca, String metodo) {
    }

    public static List<NomeEncontrado> detectarNomesAncorados(String texto) {
        List<NomeEncontrado> encontrados = new ArrayList<>();
        Matcher m = REGEX_NOME_ANCORADO.matcher(texto);
        while (m.find()) {
            encontrados.add(new NomeEncontrado(m.group(2), "alta", "ancora_rotulo"));
        }
        return encontrados;
    }

    public record CpfEncontrado(String texto, boolean valido) {
    }

    public record ResultadoVarredura(
            List<CpfEncontrado> cpfsEncontrados, List<NomeEncontrado> nomesEncontrados,
            String textoRedigido, int totalPiiRedigido
    ) {
    }

    public static ResultadoVarredura varrerPII(String textoDocumento) {
        List<CpfEncontrado> cpfsEncontrados = new ArrayList<>();
        Matcher mCpf = REGEX_CPF.matcher(textoDocumento);
        while (mCpf.find()) {
            String texto = mCpf.group();
            cpfsEncontrados.add(new CpfEncontrado(texto, validarCPF(texto)));
        }

        List<NomeEncontrado> nomesEncontrados = detectarNomesAncorados(textoDocumento);

        String textoRedigido = textoDocumento;
        for (CpfEncontrado cpf : cpfsEncontrados) {
            if (cpf.valido()) {
                textoRedigido = textoRedigido.replace(cpf.texto(), "[CPF_REDIGIDO]");
            }
        }
        for (NomeEncontrado nome : nomesEncontrados) {
            textoRedigido = textoRedigido.replace(nome.nome(), "[NOME_REDIGIDO]");
        }

        long cpfsValidos = cpfsEncontrados.stream().filter(CpfEncontrado::valido).count();
        return new ResultadoVarredura(cpfsEncontrados, nomesEncontrados, textoRedigido, (int) cpfsValidos + nomesEncontrados.size());
    }
}
