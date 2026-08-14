package com.trialforge.gateway;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Multi-Index (Modulo 4.1): um indice por dominio de agente do TrialForge —
 * ICF, Protocolo, CSR — em vez de um banco unico e achatado. Cada dominio tem
 * vocabulario e fonte regulatoria proprios (Modulo 3.1: e o mesmo motivo que
 * justifica agentes especialistas), entao misturar os tres num indice so so
 * adicionaria ruido de recuperacao sem necessidade.
 */
public final class Indices {

    private Indices() {
    }

    /** Ordem fixa dos indices — usada pra iteracao deterministica (empates em
     * buscarEmTodosIndices favorecem o primeiro da lista), equivalente a
     * Object.keys(INDICES) em JS / dict de insercao em Python. */
    public static final List<String> NOMES_INDICES = List.of("icf", "protocolo", "csr");

    public static final Map<String, List<Clausula>> INDICES = criarIndices();

    // Roteamento pro indice certo (Modulo 4.1) reaproveita a mesma
    // classificacao de intencao que ja decide o Model Router (Modulo 4.2).
    public static final Map<String, String> INDICE_POR_INTENCAO = Map.of(
            "consulta_icf", "icf",
            "consulta_protocolo", "protocolo",
            "sintese_csr", "csr");

    private static Map<String, List<Clausula>> criarIndices() {
        Map<String, List<Clausula>> indices = new LinkedHashMap<>();

        indices.put("icf", List.of(
                new Clausula(
                        "Assentimento de menores de idade em estudos clínicos",
                        "Para participantes entre 12 e 17 anos, é necessário assentimento por escrito, "
                                + "além do consentimento do responsável legal (RDC ANVISA 466/2012, Art. 4º).",
                        "RDC ANVISA 466/2012, Art. 4º"),
                new Clausula(
                        "Direito de retirada do participante do estudo a qualquer momento",
                        "O participante pode retirar seu consentimento a qualquer momento, sem necessidade "
                                + "de justificativa e sem prejuízo ao seu tratamento (RDC ANVISA 466/2012, Art. 5º).",
                        "RDC ANVISA 466/2012, Art. 5º")));

        // "doze anos" aqui e o mesmo numero da emenda etica do Modulo 3.4/3.5
        // (idade minima 13 -> 12) — o mesmo criterio de protocolo revisitado
        // em outra camada.
        indices.put("protocolo", List.of(
                new Clausula(
                        "Critério de idade mínima para inclusão no estudo",
                        "A idade mínima para participação no estudo é de doze anos completos na data "
                                + "da assinatura do assentimento, conforme a versão vigente do protocolo aprovada "
                                + "pelo comitê de ética.",
                        "Protocolo Clínico TrialForge, critério de inclusão nº 2"),
                new Clausula(
                        "Critério de exclusão por interação medicamentosa",
                        "Participantes em uso concomitante de medicação com interação farmacológica "
                                + "documentada são excluídos do estudo, conforme a seção 4.2 do protocolo.",
                        "Protocolo Clínico TrialForge, critério de exclusão nº 4")));

        indices.put("csr", List.of(
                new Clausula(
                        "Apresentação de eventos adversos no relatório final",
                        "Eventos adversos devem ser apresentados por gravidade e causalidade, seguindo "
                                + "a estrutura de seções do guia ICH E3, sem agregação que oculte eventos "
                                + "individuais graves.",
                        "ICH E3, seção 12"),
                new Clausula(
                        "Significância estatística do desfecho primário",
                        "O desfecho primário só pode ser reportado como positivo se atingir o nível de "
                                + "significância pré-especificado no plano de análise estatística, sem ajuste "
                                + "post-hoc.",
                        "ICH E3, seção 11; Plano de Análise Estatística do estudo")));

        return indices;
    }
}
