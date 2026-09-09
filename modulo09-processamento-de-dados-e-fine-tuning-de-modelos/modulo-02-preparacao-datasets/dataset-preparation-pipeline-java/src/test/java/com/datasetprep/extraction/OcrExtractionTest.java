package com.datasetprep.extraction;

import org.junit.jupiter.api.Test;

import java.util.HashMap;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class OcrExtractionTest {

    @Test
    void parseiaValorFormatoBrasileiroComMilhar() {
        assertThat(OcrExtraction.parsearValorBRL("3.210,50")).isEqualTo(3210.5);
    }

    @Test
    void parseiaValorComPrefixoRS() {
        assertThat(OcrExtraction.parsearValorBRL("R$ 145,90")).isEqualTo(145.9);
    }

    @Test
    void parseiaOrcamentoAutoDeTextoOcrTolerandoRotuloVariavel() {
        String textoOcr = "ATIVA ORCAMENTOS AUTOMOTIVOS OFICINA ESTRELA LTDA\nSegurado: Marcos Vinicius Andrade Pereira\n"
                + "Placa do veiculo: QJK-4F82\nValor total do reparo: R$ 3.210,50";
        var campos = OcrExtraction.parsearOrcamentoAuto(textoOcr);
        assertThat(campos.segurado()).isEqualTo("Marcos Vinicius Andrade Pereira");
        assertThat(campos.placa()).isEqualTo("QJK-4F82");
        assertThat(campos.valor()).isEqualTo(3210.50);
    }

    @Test
    void parseiaReciboSaudeDeTextoOcr() {
        String textoOcr = "CLINICA VITALIS SAUDE OCUPACIONAL\nBeneficiario: Carlos Eduardo Martins\n"
                + "Procedimento: Consulta Cardiologica\nValor cobrado: R$ 380,00";
        var campos = OcrExtraction.parsearReciboSaude(textoOcr);
        assertThat(campos.beneficiario()).isEqualTo("Carlos Eduardo Martins");
        assertThat(campos.procedimento()).isEqualTo("Consulta Cardiologica");
        assertThat(campos.valor()).isEqualTo(380.00);
    }

    @Test
    void exemploSemCampoObrigatorioEMarcadoInvalido() {
        Map<String, Object> saida = new HashMap<>();
        saida.put("segurado", "Fulano");
        saida.put("placa", null);
        saida.put("valor", 100.0);
        var v = OcrExtraction.validarExemplo("amplitude-auto", saida);
        assertThat(v.valido()).isFalse();
        assertThat(v.erros()).anyMatch(e -> e.contains("placa"));
    }

    @Test
    void exemploComValorZeroOuNegativoEMarcadoInvalido() {
        Map<String, Object> saida = new HashMap<>();
        saida.put("segurado", "Fulano");
        saida.put("placa", "ABC-1234");
        saida.put("valor", 0.0);
        var v = OcrExtraction.validarExemplo("amplitude-auto", saida);
        assertThat(v.valido()).isFalse();
    }
}
