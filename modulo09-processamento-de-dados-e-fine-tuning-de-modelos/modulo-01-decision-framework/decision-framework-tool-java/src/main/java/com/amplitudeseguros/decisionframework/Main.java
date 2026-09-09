package com.amplitudeseguros.decisionframework;

import com.amplitudeseguros.decisionframework.config.AmplitudeConfig;
import com.amplitudeseguros.decisionframework.config.Caso;
import com.amplitudeseguros.decisionframework.config.ConfigLoader;

import java.util.List;
import java.util.Locale;

/**
 * Demo: aplica o framework de decisão aos 3 casos reais da Amplitude Seguros
 * (Auto aprovado, Saúde Empresarial reprovado só por dado -> Real Options,
 * Atendimento ao Cliente reprovado por natureza da tarefa) -- equivalente ao
 * fluxo principal de decision-framework-tool.js.
 */
public final class Main {

    public static void main(String[] args) {
        AmplitudeConfig config = ConfigLoader.carregarConfiguracao();
        double[] pesosAHP = Ahp.derivarPesos(config.ahp().matriz());
        Ahp.Consistencia consistencia = Ahp.calcularConsistencia(config.ahp().matriz(), pesosAHP);

        System.out.println("===== Demo: Framework de 4 Perguntas -- versão de análise de decisão financeira =====");
        System.out.printf(Locale.ROOT, "%nPesos derivados por AHP: p1=%.3f  p2=%.3f  p3=%.3f  p4=%.3f%n",
                pesosAHP[0], pesosAHP[1], pesosAHP[2], pesosAHP[3]);
        System.out.printf(Locale.ROOT, "Consistência do julgamento: lambda_max=%.4f  CI=%.4f  CR=%.4f (%s)%n",
                consistencia.lambdaMax(), consistencia.ci(), consistencia.cr(),
                consistencia.consistente() ? "consistente, CR < 0.10" : "INCONSISTENTE");

        Caso auto = config.caso("amplitude-auto");
        Caso saude = config.caso("amplitude-saude-empresarial");
        Caso atendimento = config.caso("amplitude-atendimento-cliente");

        imprimirCasoAprovado(auto, FrameworkEvaluation.avaliarFramework(auto.scores(), pesosAHP, config.limiarVerde()));
        imprimirCasoReprovadoPorDado(saude,
                FrameworkEvaluation.avaliarFramework(saude.scores(), pesosAHP, config.limiarVerde()), config.limiarVerde());
        imprimirCasoReprovadoGeral(atendimento,
                FrameworkEvaluation.avaliarFramework(atendimento.scores(), pesosAHP, config.limiarVerde()));

        System.out.println();
        System.out.println("Três casos, três respostas diferentes. Auto: sim. Saúde Empresarial: ainda não,");
        System.out.println("só falta dado, e dado é questão de tempo. Atendimento ao Cliente: não -- a tarefa");
        System.out.println("em si é aberta e instável demais. Mais dado não resolve um problema que não é de dado.");
    }

    private static void imprimirGovernanca(Caso caso) {
        GovernanceGate.Resultado g = GovernanceGate.validar(caso.governanca());
        String detalhe = caso.governanca().dadoSensivelLGPD()
                ? "dado de categoria sensível (LGPD Art. 5º, II) -- DPA verificado."
                : "dado pessoal comum, não sensível.";
        System.out.printf("  [Governança] %s -- %s Base legal: %s%n",
                g.aprovado() ? "APROVADO" : "BLOQUEADO", detalhe, caso.governanca().baseLegalDescricao());
    }

    private static void imprimirPerguntas(FrameworkEvaluation.Resultado r) {
        for (int i = 0; i < FrameworkEvaluation.CHAVES_PERGUNTAS.size(); i++) {
            String chave = FrameworkEvaluation.CHAVES_PERGUNTAS.get(i);
            FrameworkEvaluation.SinalPergunta s = r.sinaisPorPergunta().get(chave);
            System.out.printf(Locale.ROOT, "  Pergunta %d [%s, score %.2f]%n", i + 1, s.sinal(), s.score());
        }
    }

    private static void imprimirCasoAprovado(Caso caso, FrameworkEvaluation.Resultado resultado) {
        System.out.printf("%n===== %s =====%n", caso.nome());
        System.out.printf("Tarefa: %s%n%n", caso.tarefa());
        imprimirGovernanca(caso);
        imprimirPerguntas(resultado);
        System.out.printf(Locale.ROOT, "%n  Score composto (AHP): %.2f%n", resultado.scoreComposto());
        System.out.printf("  Recomendação (gate): %s%n", resultado.recomendacao());

        NpvCalculator.Resultado npv = NpvCalculator.calcularNPV(NpvCalculator.paramsDeterministicos(caso.financeiro()));
        System.out.println("  --- Análise financeira (DCF) ---");
        System.out.printf(Locale.ROOT, "  NPV em %d meses (cenário mais provável): R$ %.2f%n",
                caso.financeiro().horizonteMeses(), npv.npv());
        System.out.printf("  Breakeven: %s%n", npv.mesBreakeven() != null ? "mês " + npv.mesBreakeven() : "não atinge no horizonte");

        MonteCarloSimulator.Resultado mc = MonteCarloSimulator.simular(caso.financeiro(), 10000, Math::random);
        System.out.printf(Locale.ROOT, "%n  --- Monte Carlo (10.000 simulações) ---%n");
        System.out.printf(Locale.ROOT, "  NPV médio: R$ %.2f | P5: R$ %.2f | P95: R$ %.2f | Probabilidade positiva: %.1f%%%n",
                mc.media(), mc.p5(), mc.p95(), mc.probabilidadePositivo() * 100);

        List<SensitivityAnalyzer.Resultado> sens = SensitivityAnalyzer.analisar(caso.financeiro(), 0.2);
        System.out.println("  --- Sensibilidade (ranking por impacto no NPV, +/-20%) ---");
        for (int i = 0; i < sens.size(); i++) {
            System.out.printf(Locale.ROOT, "  %d. %s: amplitude de R$ %.2f%n", i + 1, sens.get(i).parametro(), sens.get(i).amplitude());
        }
    }

    private static void imprimirCasoReprovadoPorDado(Caso caso, FrameworkEvaluation.Resultado resultado, double limiarVerde) {
        System.out.printf("%n===== %s =====%n", caso.nome());
        System.out.printf("Tarefa: %s%n%n", caso.tarefa());
        imprimirGovernanca(caso);
        imprimirPerguntas(resultado);
        System.out.println("  Isso é um \"ainda não\", não um \"não\" -- elegível a análise de Real Options.");

        RealOptionsPricer.Resultado opcao = RealOptionsPricer.precificarOpcaoDeEsperar(
                caso.financeiro(), caso.scores().get("p3"), limiarVerde, 10000, Math::random);
        System.out.println("  --- Real Options: árvore binomial, decidir agora vs. esperar ---");
        System.out.printf("  Meses até o score de dado cruzar o limiar: %d%n", opcao.mesesParaEsperar());
        System.out.printf(Locale.ROOT, "  Valor de esperar (diferença): R$ %.2f%n", opcao.valorDeEsperar());
        System.out.printf("  Recomendação: %s%n", opcao.recomendacao());
        System.out.printf("  Reavaliação agendada em: %s%n", opcao.reavaliacaoAgendadaEm());
    }

    private static void imprimirCasoReprovadoGeral(Caso caso, FrameworkEvaluation.Resultado resultado) {
        System.out.printf("%n===== %s =====%n", caso.nome());
        System.out.printf("Tarefa: %s%n%n", caso.tarefa());
        imprimirGovernanca(caso);
        imprimirPerguntas(resultado);
        System.out.println("  Reparem: a pergunta 3 (dado suficiente) está VERDE. Não é falta de dado.");
        System.out.printf("  Recomendação: %s%n", resultado.recomendacao());
    }
}
