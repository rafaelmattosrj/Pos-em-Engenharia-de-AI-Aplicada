package com.datasetprep.pii;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class PiiScrubbingGateTest {

    @Test
    void cpfDeTesteConhecidoEValidoDigitoVerificadorBate() {
        assertThat(PiiScrubbingGate.validarCPF("111.444.777-35")).isTrue();
    }

    @Test
    void trocarOUltimoDigitoVerificadorInvalidaOCpf() {
        assertThat(PiiScrubbingGate.validarCPF("111.444.777-34")).isFalse();
    }

    @Test
    void sequenciaDeDigitosRepetidosNuncaECpfValido() {
        assertThat(PiiScrubbingGate.validarCPF("000.000.000-00")).isFalse();
    }

    @Test
    void cpfSemMascaraValidaIgual() {
        assertThat(PiiScrubbingGate.validarCPF("11144477735")).isTrue();
    }

    @Test
    void digitosAleatoriosSemDigitoVerificadorCoerenteERejeitado() {
        assertThat(PiiScrubbingGate.validarCPF("123.456.789-00")).isFalse();
    }

    @Test
    void extraiNomeDepoisDeSegurado() {
        var encontrados = PiiScrubbingGate.detectarNomesAncorados("Segurado: Marcos Vinicius Andrade Pereira\nPlaca: ABC-1234");
        assertThat(encontrados).hasSize(1);
        assertThat(encontrados.get(0).nome()).isEqualTo("Marcos Vinicius Andrade Pereira");
    }

    @Test
    void extraiNomeDepoisDeBeneficiarioComAcento() {
        var encontrados = PiiScrubbingGate.detectarNomesAncorados("Beneficiário: Carlos Eduardo Martins");
        assertThat(encontrados.get(0).nome()).isEqualTo("Carlos Eduardo Martins");
    }

    @Test
    void extraiNomeDepoisDeBeneficiarioSemAcento() {
        var encontrados = PiiScrubbingGate.detectarNomesAncorados("Beneficiario: Carlos Eduardo Martins");
        assertThat(encontrados.get(0).nome()).isEqualTo("Carlos Eduardo Martins");
    }

    @Test
    void textoSemRotuloConhecidoNaoGeraFalsoPositivoNaAncora() {
        var encontrados = PiiScrubbingGate.detectarNomesAncorados("OFICINA ESTRELA - ORCAMENTO N. 4471");
        assertThat(encontrados).isEmpty();
    }

    @Test
    void documentoRealTemNomeECpfRedigidosPlacaEValorPreservados() {
        String doc = "OFICINA ESTRELA - ORCAMENTO N. 4471\nSegurado: Marcos Vinicius Andrade Pereira\n"
                + "CPF: 111.444.777-35\nPlaca do veiculo: QJK-4F82\n\nValor total do reparo: R$ 3.210,50";
        var resultado = PiiScrubbingGate.varrerPII(doc);
        assertThat(resultado.nomesEncontrados()).hasSize(1);
        assertThat(resultado.cpfsEncontrados().stream().filter(PiiScrubbingGate.CpfEncontrado::valido).count()).isEqualTo(1);
        assertThat(resultado.textoRedigido()).contains("[NOME_REDIGIDO]");
        assertThat(resultado.textoRedigido()).contains("[CPF_REDIGIDO]");
        assertThat(resultado.textoRedigido()).contains("QJK-4F82");
        assertThat(resultado.textoRedigido()).contains("3.210,50");
    }

    @Test
    void cpfComFormatoCertoMasDigitoVerificadorInvalidoNaoERedigido() {
        String doc = "Protocolo interno: 123.456.789-00\nSegurado: Ana Paula Ribeiro";
        var resultado = PiiScrubbingGate.varrerPII(doc);
        assertThat(resultado.cpfsEncontrados().stream().filter(PiiScrubbingGate.CpfEncontrado::valido).count()).isZero();
        assertThat(resultado.textoRedigido()).doesNotContain("[CPF_REDIGIDO]");
    }
}
