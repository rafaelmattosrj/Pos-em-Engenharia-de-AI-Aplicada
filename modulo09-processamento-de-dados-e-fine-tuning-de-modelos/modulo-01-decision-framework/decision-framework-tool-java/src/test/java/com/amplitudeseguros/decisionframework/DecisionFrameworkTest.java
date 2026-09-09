package com.amplitudeseguros.decisionframework;

import com.amplitudeseguros.decisionframework.config.AmplitudeConfig;
import com.amplitudeseguros.decisionframework.config.Caso;
import com.amplitudeseguros.decisionframework.config.ConfigLoader;
import com.amplitudeseguros.decisionframework.config.Governanca;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.within;

class DecisionFrameworkTest {

    private static AmplitudeConfig config;
    private static double[] pesosAHP;
    private static Ahp.Consistencia consistenciaAHP;

    @BeforeAll
    static void carregarConfig() {
        config = ConfigLoader.carregarConfiguracao();
        pesosAHP = Ahp.derivarPesos(config.ahp().matriz());
        consistenciaAHP = Ahp.calcularConsistencia(config.ahp().matriz(), pesosAHP);
    }

    @Nested
    class Ahp_PesosEConsistencia {

        @Test
        void pesosSomam1() {
            double soma = 0;
            for (double p : pesosAHP) soma += p;
            assertThat(soma).isCloseTo(1.0, within(1e-9));
        }

        @Test
        void pergunta3RecebeOMaiorPeso() {
            int indiceP3 = FrameworkEvaluation.CHAVES_PERGUNTAS.indexOf("p3");
            double maiorPeso = 0;
            for (double p : pesosAHP) maiorPeso = Math.max(maiorPeso, p);
            assertThat(pesosAHP[indiceP3]).isEqualTo(maiorPeso);
        }

        @Test
        void matrizEConsistente() {
            assertThat(consistenciaAHP.consistente())
                    .as("CR = %s, esperado < 0.10", consistenciaAHP.cr())
                    .isTrue();
        }
    }

    @Nested
    class Governanca_BloqueiaAntesDoAhp {

        @Test
        void casoSemBaseLegalEBloqueadoMesmoComScoresPerfeitos() {
            Caso hipotetico = casoHipotetico(new Governanca(false, false, "", true));
            FrameworkEvaluation.Resultado r = FrameworkEvaluation.avaliarCasoCompleto(hipotetico, pesosAHP, config.limiarVerde());
            assertThat(r.bloqueadoPorGovernanca()).isTrue();
            assertThat(r.aprovado()).isFalse();
            assertThat(r.scoreComposto()).isNull();
        }

        @Test
        void dadoSensivelSemDpaEBloqueado() {
            Caso hipotetico = casoHipotetico(new Governanca(true, true, "", false));
            FrameworkEvaluation.Resultado r = FrameworkEvaluation.avaliarCasoCompleto(hipotetico, pesosAHP, config.limiarVerde());
            assertThat(r.bloqueadoPorGovernanca()).isTrue();
            assertThat(r.motivosGovernanca().get(0)).contains("DPA");
        }

        @Test
        void dadoSensivelComDpaPassaNormalmente() {
            Caso hipotetico = casoHipotetico(new Governanca(true, true, "", true));
            FrameworkEvaluation.Resultado r = FrameworkEvaluation.avaliarCasoCompleto(hipotetico, pesosAHP, config.limiarVerde());
            assertThat(r.bloqueadoPorGovernanca()).isFalse();
            assertThat(r.aprovado()).isTrue();
        }

        @Test
        void osTresCasosReaisPassamNaGovernanca() {
            for (Caso caso : config.casos()) {
                GovernanceGate.Resultado g = GovernanceGate.validar(caso.governanca());
                assertThat(g.aprovado()).as("%s deveria passar: %s", caso.nome(), g.motivos()).isTrue();
            }
        }

        @Test
        void apenasSaudeEmpresarialExigeDpaDeVerdade() {
            Caso saude = config.caso("amplitude-saude-empresarial");
            assertThat(saude.governanca().dadoSensivelLGPD()).isTrue();
            for (Caso outro : config.casos()) {
                if (!outro.id().equals(saude.id())) {
                    assertThat(outro.governanca().dadoSensivelLGPD()).isFalse();
                }
            }
        }

        private Caso casoHipotetico(Governanca governanca) {
            return new Caso("hipotetico", "Caso hipotético", "tarefa",
                    Map.of("p1", 0.9, "p2", 0.9, "p3", 0.9, "p4", 0.9), governanca, null);
        }
    }

    @Nested
    class GateDe4Perguntas {

        @Test
        void scoreExatamenteNoLimiarContaComoVerde() {
            var r = FrameworkEvaluation.avaliarFramework(
                    Map.of("p1", config.limiarVerde(), "p2", 0.9, "p3", 0.9, "p4", 0.9), pesosAHP, config.limiarVerde());
            assertThat(r.sinaisPorPergunta().get("p1").sinal()).isEqualTo("VERDE");
            assertThat(r.aprovado()).isTrue();
        }

        @Test
        void quatroComScoreAltoAprovaSemFalha() {
            var r = FrameworkEvaluation.avaliarFramework(Map.of("p1", 0.9, "p2", 0.9, "p3", 0.9, "p4", 0.9), pesosAHP, config.limiarVerde());
            assertThat(r.aprovado()).isTrue();
            assertThat(r.perguntasFalhas()).isEmpty();
        }

        @Test
        void aprovadoDeixaDecisaoDeTecnicaEmAberto() {
            var aprovado = FrameworkEvaluation.avaliarFramework(Map.of("p1", 0.9, "p2", 0.9, "p3", 0.9, "p4", 0.9), pesosAHP, config.limiarVerde());
            var reprovado = FrameworkEvaluation.avaliarFramework(Map.of("p1", 0.2, "p2", 0.9, "p3", 0.9, "p4", 0.9), pesosAHP, config.limiarVerde());
            assertThat(aprovado.decisaoTecnicaEmAberto()).isTrue();
            assertThat(reprovado.decisaoTecnicaEmAberto()).isFalse();
        }

        @Test
        void pergunta3AbaixoDoLimiarFalhaSoDado() {
            var r = FrameworkEvaluation.avaliarFramework(Map.of("p1", 0.9, "p2", 0.9, "p3", 0.2, "p4", 0.9), pesosAHP, config.limiarVerde());
            assertThat(r.aprovado()).isFalse();
            assertThat(r.perguntasFalhas()).containsExactly(3);
            assertThat(r.falhaSoDado()).isTrue();
        }

        @Test
        void pergunta1AbaixoDoLimiarNaoEFalhaSoDado() {
            var r = FrameworkEvaluation.avaliarFramework(Map.of("p1", 0.2, "p2", 0.9, "p3", 0.9, "p4", 0.9), pesosAHP, config.limiarVerde());
            assertThat(r.falhaSoDado()).isFalse();
        }
    }

    @Nested
    class Npv_Dcf {

        @Test
        void semCrescimentoBateComFormulaFechadaDeAnuidade() {
            var params = NpvCalculator.Params.semAtraso(1000, 0, 0.05, 0.02, 1000, 12, 0.01);
            var resultado = NpvCalculator.calcularNPV(params);
            double economiaPorChamada = 0.05 - 0.02;
            double fluxoMensal = 1000 * economiaPorChamada;
            double pvAnuidade = fluxoMensal * (1 - Math.pow(1.01, -12)) / 0.01;
            double npvEsperado = -1000 + pvAnuidade;
            assertThat(resultado.npv()).isCloseTo(npvEsperado, within(0.5));
        }

        @Test
        void comAtrasoNaoGeraEconomiaDuranteEspera() {
            var params = new NpvCalculator.Params(1000, 0, 0.05, 0.02, 0, 3, 0, 3);
            var resultado = NpvCalculator.calcularNPV(params);
            assertThat(resultado.npv()).isEqualTo(0.0);
        }
    }

    @Nested
    class MonteCarlo {

        @Test
        void amostragemTriangularConvergeParaMediaTeorica() {
            long[] seed = {42};
            java.util.function.DoubleSupplier rng = () -> {
                seed[0] = (seed[0] * 1103515245L + 12345L) % 2147483648L;
                return seed[0] / 2147483648.0;
            };
            double soma = 0;
            int n = 20000;
            for (int i = 0; i < n; i++) {
                soma += MonteCarloSimulator.amostrarTriangular(0.01, 0.03, 0.05, rng);
            }
            double media = soma / n;
            double mediaTeorica = (0.01 + 0.03 + 0.05) / 3;
            assertThat(media).isCloseTo(mediaTeorica, within(0.002));
        }

        @Test
        void casoAutoTemProbabilidadeAltaDeNpvPositivo() {
            Caso auto = config.caso("amplitude-auto");
            var mc = MonteCarloSimulator.simular(auto.financeiro(), 2000, Math::random);
            assertThat(mc.probabilidadePositivo()).isGreaterThan(0.9);
        }
    }

    @Nested
    class RealOptions {

        @Test
        void esperarTemValorPositivoQuandoCustoDeErroEAlto() {
            Caso saude = config.caso("amplitude-saude-empresarial");
            var opcao = RealOptionsPricer.precificarOpcaoDeEsperar(
                    saude.financeiro(), saude.scores().get("p3"), config.limiarVerde(), 2000, Math::random);
            assertThat(opcao.mesesParaEsperar()).isGreaterThan(0);
            assertThat(opcao.recomendacao()).isEqualTo(Recomendacao.ESPERAR);
            assertThat(opcao.valorDeEsperar()).isGreaterThan(0);
        }

        @Test
        void volatilidadePositivaEFatoresUDConsistentes() {
            Caso saude = config.caso("amplitude-saude-empresarial");
            var opcao = RealOptionsPricer.precificarOpcaoDeEsperar(
                    saude.financeiro(), saude.scores().get("p3"), config.limiarVerde(), 2000, Math::random);
            assertThat(opcao.sigma()).isGreaterThan(0);
            assertThat(opcao.u() * opcao.d()).isCloseTo(1.0, within(1e-3));
            assertThat(opcao.probabilidadeRiscoNeutra()).isBetween(0.0, 1.0);
        }

        @Test
        void esperarVemComReavaliacaoAgendada() {
            Caso saude = config.caso("amplitude-saude-empresarial");
            var opcao = RealOptionsPricer.precificarOpcaoDeEsperar(
                    saude.financeiro(), saude.scores().get("p3"), config.limiarVerde(), 2000, Math::random);
            assertThat(opcao.reavaliacaoAgendadaEm()).isEqualTo("Módulo 3.2");
        }
    }

    @Nested
    class Sensibilidade {

        @Test
        void retornaOs4ParametrosRanqueadosPorAmplitudeDecrescente() {
            Caso auto = config.caso("amplitude-auto");
            List<SensitivityAnalyzer.Resultado> sens = SensitivityAnalyzer.analisar(auto.financeiro(), 0.2);
            assertThat(sens).hasSize(4);
            for (int i = 1; i < sens.size(); i++) {
                assertThat(sens.get(i - 1).amplitude()).isGreaterThanOrEqualTo(sens.get(i).amplitude());
            }
        }
    }

    @Nested
    class AplicacaoAosTresCasosReais {

        @Test
        void amplitudeAutoEAprovado() {
            Caso auto = config.caso("amplitude-auto");
            var r = FrameworkEvaluation.avaliarFramework(auto.scores(), pesosAHP, config.limiarVerde());
            assertThat(r.aprovado()).isTrue();
        }

        @Test
        void amplitudeSaudeReprovadoSoPorDado() {
            Caso saude = config.caso("amplitude-saude-empresarial");
            var r = FrameworkEvaluation.avaliarFramework(saude.scores(), pesosAHP, config.limiarVerde());
            assertThat(r.aprovado()).isFalse();
            assertThat(r.falhaSoDado()).isTrue();
        }

        @Test
        void amplitudeAtendimentoReprovadoEmP1EP4() {
            Caso atendimento = config.caso("amplitude-atendimento-cliente");
            var r = FrameworkEvaluation.avaliarFramework(atendimento.scores(), pesosAHP, config.limiarVerde());
            assertThat(r.aprovado()).isFalse();
            assertThat(r.perguntasFalhas()).containsExactly(1, 4);
            assertThat(r.falhaSoDado()).isFalse();
            assertThat(r.sinaisPorPergunta().get("p3").sinal()).isEqualTo("VERDE");
            assertThat(r.recomendacao()).isEqualTo(Recomendacao.CONTINUAR_PROMPT_RAG);
        }
    }

    @Nested
    class AhpDeComite {

        @Test
        void comiteNaoMudaOVereditoDosTresCasosReais() {
            double[][] matrizGestorProduto = {
                    {1, 1, 1.0 / 3, 0.5}, {1, 1, 1.0 / 3, 0.5}, {3, 3, 1, 2}, {2, 2, 0.5, 1}
            };
            double[][] matrizCompliance = {
                    {1, 2, 1.0 / 5, 1.0 / 3}, {0.5, 1, 1.0 / 5, 1.0 / 3}, {5, 5, 1, 3}, {3, 3, 1.0 / 3, 1}
            };
            double[][] matrizEngenharia = {
                    {1, 1, 0.5, 1}, {1, 1, 0.5, 1}, {2, 2, 1, 2}, {1, 1, 0.5, 1}
            };
            double[][] matrizAgregada = Ahp.agregarMatrizesComite(List.of(matrizGestorProduto, matrizCompliance, matrizEngenharia));
            double[] pesosComite = Ahp.derivarPesos(matrizAgregada);

            for (Caso caso : List.of(
                    config.caso("amplitude-auto"), config.caso("amplitude-saude-empresarial"),
                    config.caso("amplitude-atendimento-cliente"))) {
                boolean original = FrameworkEvaluation.avaliarFramework(caso.scores(), pesosAHP, config.limiarVerde()).aprovado();
                boolean comite = FrameworkEvaluation.avaliarFramework(caso.scores(), pesosComite, config.limiarVerde()).aprovado();
                assertThat(comite).as("veredito de %s mudou entre matriz única e comitê", caso.id()).isEqualTo(original);
            }
        }

        @Test
        void comiteDeUmAvaliadorReproduzOsPesosOriginais() {
            double[][] matrizSolo = {{1, 1, 1.0 / 3, 0.5}, {1, 1, 1.0 / 3, 0.5}, {3, 3, 1, 2}, {2, 2, 0.5, 1}};
            double[][] agregada = Ahp.agregarMatrizesComite(List.<double[][]>of(matrizSolo));
            double[] pesosSolo = Ahp.derivarPesos(agregada);
            double[] pesosOriginais = Ahp.derivarPesos(matrizSolo);
            for (int i = 0; i < pesosSolo.length; i++) {
                assertThat(pesosSolo[i]).isCloseTo(pesosOriginais[i], within(1e-9));
            }
        }

        @Test
        void matrizAgregadaEReciprocamenteValida() {
            double[][] m1 = {{1, 1, 1.0 / 3, 0.5}, {1, 1, 1.0 / 3, 0.5}, {3, 3, 1, 2}, {2, 2, 0.5, 1}};
            double[][] m2 = {{1, 2, 1.0 / 5, 1.0 / 3}, {0.5, 1, 1.0 / 5, 1.0 / 3}, {5, 5, 1, 3}, {3, 3, 1.0 / 3, 1}};
            double[][] agregada = Ahp.agregarMatrizesComite(List.of(m1, m2));
            for (int i = 0; i < agregada.length; i++) {
                for (int j = 0; j < agregada.length; j++) {
                    assertThat(agregada[i][j] * agregada[j][i]).isCloseTo(1.0, within(1e-9));
                }
            }
        }
    }
}
