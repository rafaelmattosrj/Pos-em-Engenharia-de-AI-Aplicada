package com.iadeva.agent;

import com.iadeva.agent.core.*;
import com.iadeva.agent.model.AgentTrace;
import com.iadeva.agent.model.PlanDecision;
import com.iadeva.agent.model.ToolResult;
import com.iadeva.agent.observability.AgentTelemetry;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

/**
 * Testes unitários do Agent Loop — valida os contratos de comportamento do agente.
 *
 * Cenários cobertos:
 * 1. AgentLoop respeita o limite de maxSteps configurado no contrato
 * 2. CircuitBreaker interrompe o loop após 3 respostas inválidas consecutivas
 * 3. Executor retorna ToolResult de erro para tool desconhecida
 */
@ExtendWith(MockitoExtension.class)
class AgentLoopTest {

    @Mock
    private Planner planner;

    @Mock
    private Executor executor;

    @Mock
    private AgentTelemetry telemetry;

    // CircuitBreaker é instanciado diretamente (sem mock) nos cenários 1 e 3
    // para testar comportamento real no cenário 2

    private AgentContract contract;

    @BeforeEach
    void setup() {
        contract = new AgentContract();
    }

    // -------------------------------------------------------------------------
    // Cenário 1: AgentLoop respeita maxSteps
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("AgentLoop deve parar ao atingir o limite de maxSteps configurado no contrato")
    void agentLoop_deveRespeitar_maxSteps() {
        // Arrange: contrato com maxSteps = 3
        // O planner sempre retorna uma PlanDecision válida com done=false
        PlanDecision planContinue = new PlanDecision(
                "Analisando métricas",
                "Buscar métricas do serviço",
                "getMetrics",
                Map.of("service", "api-gateway"),
                "Obter p99 < 1000ms",
                false // nunca sinaliza done
        );

        when(planner.plan(anyString(), any(AgentState.class), any(AgentContract.class)))
                .thenReturn(planContinue);

        when(executor.execute(any(PlanDecision.class)))
                .thenReturn(ToolResult.ok("getMetrics", "{\"p99\": 1800}"));

        // Configura telemetria para retornar trace finalizado
        AgentTrace expectedTrace = new AgentTrace("teste maxSteps");
        expectedTrace.setTotalSteps(3);
        expectedTrace.setFinalResult("Execução interrompida após 3 steps");
        when(telemetry.finalize(any(AgentState.class), anyLong())).thenReturn(expectedTrace);

        CircuitBreaker circuitBreaker = new CircuitBreaker();
        AgentLoop agentLoop = new AgentLoop(planner, executor, circuitBreaker, telemetry, contract);

        // Sobrescreve o contrato para maxSteps = 3 via reflection (simulando configuração)
        // Como não há setter, usamos um contrato customizado via subclasse anônima
        AgentContract contractWith3Steps = new AgentContract() {
            @Override
            public int getMaxSteps() { return 3; }
            @Override
            public int getMaxTimeSeconds() { return 120; }
            @Override
            public int getNoProgressSteps() { return 10; } // desativa detecção de no-progress
        };

        AgentLoop agentLoopLimited = new AgentLoop(planner, executor, circuitBreaker, telemetry, contractWith3Steps);

        // Act
        AgentTrace trace = agentLoopLimited.run("Diagnosticar degradação no api-gateway");

        // Assert
        // Planner deve ser chamado exatamente maxSteps vezes (3)
        verify(planner, times(3)).plan(anyString(), any(AgentState.class), any(AgentContract.class));

        // O trace deve indicar que foi limitado pelo maxSteps
        assertThat(trace).isNotNull();
        assertThat(trace.getFinalResult()).contains("3 steps");
    }

    // -------------------------------------------------------------------------
    // Cenário 2: CircuitBreaker quebra após 3 respostas inválidas consecutivas
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("CircuitBreaker deve abrir após 3 registros de respostas inválidas consecutivas")
    void circuitBreaker_deveQuebrar_apos3InvalidasConsecutivas() {
        // Arrange
        CircuitBreaker circuitBreaker = new CircuitBreaker();

        // Act: registra 2 inválidas — ainda não deve quebrar
        circuitBreaker.recordInvalid();
        circuitBreaker.recordInvalid();

        assertThat(circuitBreaker.shouldBreak())
                .as("CircuitBreaker NÃO deve abrir antes de atingir o threshold")
                .isFalse();

        // Registra a 3ª inválida — deve quebrar agora
        circuitBreaker.recordInvalid();

        assertThat(circuitBreaker.shouldBreak())
                .as("CircuitBreaker DEVE abrir após 3 respostas inválidas consecutivas")
                .isTrue();

        assertThat(circuitBreaker.getInvalidCount()).isEqualTo(3);

        // Verifica que uma resposta válida reseta o contador
        circuitBreaker.recordValid();

        assertThat(circuitBreaker.shouldBreak())
                .as("CircuitBreaker deve ser resetado após resposta válida")
                .isFalse();

        assertThat(circuitBreaker.getInvalidCount()).isZero();
    }

    // -------------------------------------------------------------------------
    // Cenário 3: Executor retorna erro para tool desconhecida
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("Executor deve retornar ToolResult de erro quando a tool solicitada não existe no registry")
    void executor_deveRetornarErro_paraToolDesconhecida() {
        // Arrange: Executor real com lista de tools vazia (nenhuma tool registrada)
        Executor executorSemTools = new Executor(java.util.List.of());

        PlanDecision planComToolInexistente = new PlanDecision(
                "Tentando usar tool inexistente",
                "Executar ferramenta desconhecida",
                "toolQueNaoExiste",
                Map.of("param", "valor"),
                "Verificar funcionamento",
                false
        );

        // Act
        ToolResult result = executorSemTools.execute(planComToolInexistente);

        // Assert
        assertThat(result.success())
                .as("Deve retornar sucesso=false para tool desconhecida")
                .isFalse();

        assertThat(result.toolName())
                .as("Deve preservar o nome da tool no resultado")
                .isEqualTo("toolQueNaoExiste");

        assertThat(result.error())
                .as("Mensagem de erro deve indicar que a tool não foi encontrada")
                .contains("toolQueNaoExiste");

        assertThat(result.output())
                .as("Output deve ser null em caso de erro")
                .isNull();
    }
}
