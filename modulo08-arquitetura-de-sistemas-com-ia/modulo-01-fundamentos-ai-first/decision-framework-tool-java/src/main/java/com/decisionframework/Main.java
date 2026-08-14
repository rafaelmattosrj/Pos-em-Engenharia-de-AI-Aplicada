package com.decisionframework;

import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.util.List;

/**
 * Demo: imprime os 4 casos do Bloco 1 e a decomposicao da tarefa hibrida do Bloco 2
 * de forma legivel no terminal — mesmo espirito da versao JS/Python original: mostrar
 * o mecanismo funcionando, nao so os testes passando silenciosamente.
 *
 * <p>Diferenca deliberada em relacao ao original: la, {@code rodarTestes()}/
 * {@code executar_demo()} sao chamados na mesma execucao direta do arquivo (node/
 * python3), porque JS e Python nao separam "codigo de teste" de "codigo de producao"
 * em arquivos/fases distintas por padrao. Em Java, a convencao do projeto (mesma dos
 * demais portes do repositorio) e separar testes automatizados (JUnit 5, rodados via
 * {@code mvn test}) do ponto de entrada de demo ({@code Main}, rodado via
 * {@code mvn exec:java}). O comportamento observavel da demo em si e identico ao
 * original; so a orquestracao teste+demo-no-mesmo-comando não foi replicada.</p>
 */
public final class Main {

    private Main() {
    }

    public static void main(String[] args) {
        // Forca UTF-8 na saida, independente do charset padrao da JVM/console (no
        // Windows, o console costuma usar cp1252/cp850 por padrao em Java 17, o que
        // corrompe os acentos do texto original em portugues). Isso nao muda o
        // comportamento observavel do conteudo, so garante que os bytes emitidos
        // sejam sempre UTF-8; se o terminal nao estiver configurado para UTF-8
        // (ex.: `chcp 65001` no Windows), a renderizacao visual pode variar, mas os
        // bytes escritos estarao corretos.
        System.setOut(new PrintStream(System.out, true, StandardCharsets.UTF_8));

        System.out.println("=".repeat(72));
        System.out.println("Bloco 1: as 4 combinações da árvore pura de três perguntas");
        System.out.println("=".repeat(72));

        Object[][] casosBloco1 = {
                {"P1=True,  P2=False, P3=False", true, false, false},
                {"P1=False, P2=True,  P3=False", false, true, false},
                {"P1=False, P2=False, P3=True ", false, false, true},
                {"P1=False, P2=False, P3=False", false, false, false},
        };
        for (Object[] caso : casosBloco1) {
            String rotulo = (String) caso[0];
            boolean p1 = (boolean) caso[1];
            boolean p2 = (boolean) caso[2];
            boolean p3 = (boolean) caso[3];
            Classificacao classificacao = DecisionFramework.classificarTarefa(p1, p2, p3);
            System.out.println("  " + rotulo + " -> " + classificacao);
        }

        System.out.println();
        System.out.println("=".repeat(72));
        System.out.println("Bloco 2: decompondo a tarefa híbrida de referência");
        System.out.println("(TrialForge - Emenda de Protocolo)");
        System.out.println("=".repeat(72));

        List<Subtarefa> subtarefas = List.of(
                new Subtarefa("Extrair o que mudou entre versões do protocolo", "Extração/Interpretação",
                        false, false, true),
                new Subtarefa("Classificar o tipo de emenda (administrativa/substancial)", "Decisão de Negócio",
                        true, false, false));

        for (SubtarefaClassificada subtarefa : DecisionFramework.decomporTarefaHibrida(subtarefas)) {
            System.out.println("  Subtarefa: " + subtarefa.nome());
            System.out.println("    Tipo: " + subtarefa.tipo());
            System.out.println("    P1=" + subtarefa.p1() + ", P2=" + subtarefa.p2() + ", P3=" + subtarefa.p3());
            System.out.println("    Classificação: " + subtarefa.classificacao());
            System.out.println();
        }

        System.out.println(
                "  Nota: as outras duas linhas da tabela de referência do checklist\n"
                        + "  ('Rotear pela criticidade' e 'Regenerar documentos afetados') não\n"
                        + "  entram nesta demo de propósito, pelo mesmo motivo documentado em\n"
                        + "  DecisionFrameworkTest: misturam regra e gate condicional de um\n"
                        + "  jeito que não mapeia limpo pra uma única resposta P1/P2/P3.");
    }
}
