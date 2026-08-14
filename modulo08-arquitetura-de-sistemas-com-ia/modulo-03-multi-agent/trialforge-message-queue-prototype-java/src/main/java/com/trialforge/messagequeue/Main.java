package com.trialforge.messagequeue;

import java.io.PrintStream;
import java.nio.charset.StandardCharsets;

/**
 * Porte de main() (JS/Python) — demonstracao narrada, igual ao Slide 3 +
 * paragrafos 68-73 do TP. A suite de "testes automatizados" que o original
 * roda dentro do main() virou testes JUnit 5 idiomaticos em
 * {@code TrialForgeFluxoTest} (rode com {@code mvn test}); este Main roda so
 * a demonstracao dos 5 cenarios.
 *
 * <p>Uso: {@code mvn exec:java} (ou execute a classe compilada diretamente).</p>
 */
public final class Main {

    private Main() {
    }

    public static void main(String[] args) throws InterruptedException {
        // Windows por padrao nao usa UTF-8 no console; forcamos aqui para os
        // acentos do texto narrado saírem corretos independente do codepage do terminal.
        System.setOut(new PrintStream(System.out, true, StandardCharsets.UTF_8));

        System.out.println("===== TRIALFORGE: FILA DE MENSAGENS (Barramento sobre ExecutorService) =====\n");

        System.out.println("[Cenário 1] Caminho feliz — Protocolo primeiro, depois ICF+CSR em paralelo:");
        ResultadoFluxo feliz = TrialForgeFluxo.rodarFluxo("Estudo fase II, 12-17 anos");
        System.out.println("  Evento protocolo:pronto: " + feliz.getDadoProtocolo());
        System.out.println("  Resultado ICF: " + feliz.getResultadoICF());
        System.out.println("  Resultado CSR: " + feliz.getResultadoCSR());
        System.out.println("  Decisão do Supervisor: " + feliz.getDecisaoSupervisor());
        System.out.println("  Ordem real de execução: " + String.join(" -> ", feliz.getOrdemDeExecucao()));
        System.out.println();

        System.out.println("[Cenário 2] ICF falha, usando estratégia PROMISE_ALL (o bug do parágrafo 68-71):");
        ResultadoFluxo bug = TrialForgeFluxo.rodarFluxo("Estudo fase II, 12-17 anos",
                OpcoesFluxo.padrao().comEstrategia(Estrategia.PROMISE_ALL).forcarFalhaICF());
        System.out.println("  ok: " + bug.isOk() + " | resultado do CSR preservado? " + bug.isResultadoCSRPreservado());
        System.out.println("  -> O CSR terminou bem, mas a estratégia PROMISE_ALL descartou o lote inteiro: o resultado dele se perdeu.");
        System.out.println();

        System.out.println("[Cenário 3] Mesmo cenário, corrigido com PROMISE_ALL_SETTLED — CAP + idempotência (parágrafo 72-73):");
        ResultadoFluxo corrigido = TrialForgeFluxo.rodarFluxo("Estudo fase II, 12-17 anos",
                OpcoesFluxo.padrao().comEstrategia(Estrategia.PROMISE_ALL_SETTLED).forcarFalhaICF());
        System.out.println("  Resultado do CSR preservado? " + corrigido.isResultadoCSRPreservado());
        System.out.println("  Decisão do Supervisor: " + corrigido.getDecisaoSupervisor());
        System.out.println("  Documentos de ICF gravados: " + corrigido.getDocumentosGeradosSize()
                + " (" + corrigido.getTentativasICF() + " tentativas registradas nesse único documento)");
        System.out.println("  -> O Supervisor não só decidiu o que refazer: refez de verdade, e o registro idempotente");
        System.out.println("     garantiu que a segunda tentativa não gerou um segundo documento de ICF.");
        System.out.println();

        System.out.println("[Cenário 4] Emenda do comitê de ética chega no meio do caminho — Saga de verdade:");
        ResultadoFluxo divergencia = TrialForgeFluxo.rodarFluxo("Estudo com emenda ética",
                OpcoesFluxo.padrao().comEmendaEtica());
        System.out.println("  Verificação de consistência: " + divergencia.getVerificacao());
        System.out.println("  Decisão do Supervisor: " + divergencia.getDecisaoSupervisor());
        System.out.println("  Resultado final do ICF (nunca tocado): " + divergencia.getResultadoICF().secao());
        System.out.println("  Resultado final do CSR (regenerado): " + divergencia.getResultadoCSR().sintese());
        System.out.println("  Histórico de revisões do protocolo: " + divergencia.getEstadoProtocolo().getHistorico());
        System.out.println("  -> O ICF terminou ANTES da emenda chegar — seu resultado já nasceu correto, nunca precisou");
        System.out.println("     ser refeito. O CSR terminou DEPOIS da emenda, mas com o critério antigo — o Supervisor");
        System.out.println("     detectou e regenerou só a parte dele: compensação Saga, desfazer só o que precisa.");
        System.out.println();

        System.out.println("[Cenário 5] CAP completo — timeout real, retry com limite, e o limite esgotando:");
        ResultadoFluxo travado = TrialForgeFluxo.rodarFluxo("Estudo com agente travado",
                OpcoesFluxo.padrao().forcarTravamentoICF());
        System.out.println("  Erro do despacho original: " + travado.getErroICF());
        System.out.println("  Decisão do Supervisor: " + travado.getDecisaoSupervisor()
                + " (" + travado.getTentativasRetry() + " tentativa(s) de retry)");
        System.out.println("  -> O ICF travou de verdade (corrida real contra o timeout), não lançou uma exceção —");
        System.out.println("     e o retry resolveu na primeira tentativa, porque a causa raiz era transitória.");
        System.out.println();

        ResultadoFluxo esgotado = TrialForgeFluxo.rodarFluxo("Estudo com falha persistente",
                OpcoesFluxo.padrao().forcarTravamentoICF().persistirFalhaNoRetry());
        System.out.println("  Agora com falha PERSISTENTE (não transitória):");
        System.out.println("  Decisão do Supervisor: " + esgotado.getDecisaoSupervisor()
                + " (" + esgotado.getTentativasRetry() + " tentativa(s) de retry)");
        System.out.println("  Reação final ok? " + esgotado.isOk());
        System.out.println("  -> " + TrialForgeFluxo.MAX_TENTATIVAS + " tentativas esgotadas, o Supervisor desiste do ICF");
        System.out.println("     e segue em frente sem esse resultado — Disponibilidade sobre Consistência (Teorema CAP).");
        System.out.println();

        System.exit(0);
    }
}
