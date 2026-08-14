package com.decisionframework;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Cobre os mesmos cenarios dos testes automatizados originais
 * (decision-framework-tool.js#rodarTestes / decision_framework_tool.py#TestClassificarTarefa
 * e TestDecomporTarefaHibrida): as 4 combinacoes da arvore pura (Bloco 1) e a
 * decomposicao da tarefa hibrida de referencia do TrialForge (Bloco 2).
 */
class DecisionFrameworkTest {

    @Nested
    @DisplayName("Bloco 1: árvore pura (4 combinações do checklist)")
    class ClassificarTarefaTest {

        @ParameterizedTest(name = "p1={0}, p2={1}, p3={2} -> Regra determinística")
        @CsvSource({
                "true, true,  true",
                "true, true,  false",
                "true, false, true",
                "true, false, false",
        })
        @DisplayName("p1=true com qualquer p2/p3 -> Regra determinística")
        void p1VerdadeiroComQualquerP2P3EhRegraDeterministica(boolean p1, boolean p2, boolean p3) {
            assertThat(DecisionFramework.classificarTarefa(p1, p2, p3))
                    .isEqualTo(Classificacao.REGRA_DETERMINISTICA);
        }

        @ParameterizedTest(name = "p1=false, p2=true, p3={0} -> Approval Gate")
        @CsvSource({"true", "false"})
        @DisplayName("p1=false, p2=true com qualquer p3 -> Agente com Approval Gate obrigatório")
        void p1FalsoP2VerdadeiroComQualquerP3EhApprovalGate(boolean p3) {
            assertThat(DecisionFramework.classificarTarefa(false, true, p3))
                    .isEqualTo(Classificacao.APPROVAL_GATE_OBRIGATORIO);
        }

        @Test
        @DisplayName("p1=false, p2=false, p3=true -> Agente autônomo, com observabilidade completa")
        void p1FalsoP2FalsoP3VerdadeiroEhAgenteAutonomo() {
            assertThat(DecisionFramework.classificarTarefa(false, false, true))
                    .isEqualTo(Classificacao.AGENTE_AUTONOMO);
        }

        @Test
        @DisplayName("p1=false, p2=false, p3=false -> Regra determinística (enumerável)")
        void p1FalsoP2FalsoP3FalsoEhRegraDeterministicaEnumeravel() {
            assertThat(DecisionFramework.classificarTarefa(false, false, false))
                    .isEqualTo(Classificacao.REGRA_ENUMERAVEL);
        }

        @Test
        @DisplayName("toString() de cada classificação bate com o texto exato do checklist")
        void toStringBateComTextoDoChecklist() {
            assertThat(Classificacao.REGRA_DETERMINISTICA.toString()).isEqualTo("Regra determinística");
            assertThat(Classificacao.APPROVAL_GATE_OBRIGATORIO.toString())
                    .isEqualTo("Agente com Approval Gate obrigatório");
            assertThat(Classificacao.AGENTE_AUTONOMO.toString())
                    .isEqualTo("Agente autônomo, com observabilidade completa");
            assertThat(Classificacao.REGRA_ENUMERAVEL.toString())
                    .isEqualTo("Regra determinística (mesmo parecendo complexa, se é enumerável, é regra)");
        }
    }

    @Nested
    @DisplayName("Bloco 2: decomposição de tarefa híbrida (referência TrialForge)")
    class DecomporTarefaHibridaTest {

        // Nota honesta sobre a especificacao (mesma do original): das quatro linhas
        // da tabela de referencia do checklist (secao "Referência: TrialForge -
        // Emenda de Protocolo"), so estas duas mapeiam de forma limpa para uma unica
        // resposta p1/p2/p3. "Rotear pela criticidade" e "Regenerar documentos
        // afetados" misturam regra e gate condicional de um jeito mais sutil que a
        // arvore pura de tres perguntas nao representa sozinha, por isso ficam de
        // fora do teste automatizado. Limitacao real da arvore, nao bug do porte.
        private final List<Subtarefa> subtarefasReferencia = List.of(
                new Subtarefa(
                        "Extrair o que mudou entre versões do protocolo",
                        "Extração/Interpretação",
                        // Não é regra enumerável: é extração/interpretação de linguagem
                        // natural sobre o texto do protocolo.
                        false,
                        // A extração em si não causa, diretamente, um erro caro e irreversível.
                        false,
                        // O comportamento muda bastante conforme o que de fato mudou entre
                        // as duas versões do protocolo.
                        true),
                new Subtarefa(
                        "Classificar o tipo de emenda (administrativa/substancial)",
                        "Decisão de Negócio",
                        // Segue um critério fixo definido pela ANVISA: regra finita cobre
                        // os casos reais.
                        true,
                        false,
                        false));

        @Test
        @DisplayName("\"Extrair o que mudou...\" -> Agente autônomo, com observabilidade completa")
        void extrairOQueMudouEhAgenteAutonomo() {
            List<SubtarefaClassificada> resultado = DecisionFramework.decomporTarefaHibrida(subtarefasReferencia);
            assertThat(resultado.get(0).classificacao()).isEqualTo(Classificacao.AGENTE_AUTONOMO);
        }

        @Test
        @DisplayName("\"Classificar o tipo de emenda...\" -> Regra determinística")
        void classificarTipoDeEmendaEhRegraDeterministica() {
            List<SubtarefaClassificada> resultado = DecisionFramework.decomporTarefaHibrida(subtarefasReferencia);
            assertThat(resultado.get(1).classificacao()).isEqualTo(Classificacao.REGRA_DETERMINISTICA);
        }

        @Test
        @DisplayName("decomporTarefaHibrida preserva nome, tipo e p1/p2/p3 de cada subtarefa")
        void preservaCamposOriginaisDeCadaSubtarefa() {
            List<SubtarefaClassificada> resultado = DecisionFramework.decomporTarefaHibrida(subtarefasReferencia);

            assertThat(resultado).hasSameSizeAs(subtarefasReferencia);

            assertThat(resultado.get(0).nome()).isEqualTo(subtarefasReferencia.get(0).nome());
            assertThat(resultado.get(0).tipo()).isEqualTo(subtarefasReferencia.get(0).tipo());
            assertThat(resultado.get(0).p1()).isEqualTo(subtarefasReferencia.get(0).p1());
            assertThat(resultado.get(0).p2()).isEqualTo(subtarefasReferencia.get(0).p2());
            assertThat(resultado.get(0).p3()).isEqualTo(subtarefasReferencia.get(0).p3());

            assertThat(resultado.get(1).nome()).isEqualTo(subtarefasReferencia.get(1).nome());
            assertThat(resultado.get(1).tipo()).isEqualTo(subtarefasReferencia.get(1).tipo());
        }
    }
}
