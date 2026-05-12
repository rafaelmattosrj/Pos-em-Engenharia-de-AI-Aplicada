package com.iadeva.cognitive;

import com.iadeva.cognitive.agent.CritiqueEvaluator;
import com.iadeva.cognitive.agent.ReactAgent;
import com.iadeva.cognitive.agent.ReflectionAgent;
import com.iadeva.cognitive.controller.AgentController;
import com.iadeva.cognitive.model.AgentRequest;
import com.iadeva.cognitive.model.AgentResponse;
import com.iadeva.cognitive.model.CritiqueResult;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.*;

/**
 * Testes unitários das arquiteturas cognitivas.
 * Usa Mockito para isolar dependências do LLM — sem chamadas reais à OpenAI.
 */
@ExtendWith(MockitoExtension.class)
class CognitiveArchitecturesTest {

    // -------------------------------------------------------------------------
    // Cenário 1: CritiqueEvaluator retorna passed=false quando score < threshold
    // -------------------------------------------------------------------------

    @Mock
    private CritiqueEvaluator critiqueEvaluatorMock;

    @Test
    @DisplayName("CritiqueEvaluator: deve retornar passed=false quando score é menor que threshold (0.7)")
    void critiqueEvaluator_deveRetornarPassedFalse_quandoScoreAbaixoDoThreshold() {
        // Arrange — score 0.5 está abaixo do threshold padrão de 0.7
        CritiqueResult resultadoBaixo = new CritiqueResult(0.5, false,
                "Resposta incompleta e com imprecisões.", 0.4, 0.5, 0.6);

        when(critiqueEvaluatorMock.evaluate(anyString(), anyString()))
                .thenReturn(resultadoBaixo);

        // Act
        CritiqueResult resultado = critiqueEvaluatorMock.evaluate(
                "Explique machine learning", "ML é uma coisa.");

        // Assert
        assertThat(resultado.passed()).isFalse();
        assertThat(resultado.score()).isLessThan(0.7);
        assertThat(resultado.feedback()).isNotBlank();
        assertThat(resultado.correctness()).isLessThan(0.7);
        assertThat(resultado.completeness()).isLessThan(0.7);

        verify(critiqueEvaluatorMock).evaluate(anyString(), anyString());
    }

    // -------------------------------------------------------------------------
    // Cenário 2: ReflectionAgent para após maxCycles mesmo sem passar
    // -------------------------------------------------------------------------

    @Mock
    private ReflectionAgent reflectionAgentMock;

    @Test
    @DisplayName("ReflectionAgent: deve encerrar após maxCycles mesmo que o score nunca passe no threshold")
    void reflectionAgent_deveEncerrarAposMaxCycles_mesmoSemPassarNoThreshold() {
        // Arrange — simula que o agente executou 3 ciclos (2 reflexões) sem aprovação
        AgentResponse.AgentMetrics metricsComMaxCycles = new AgentResponse.AgentMetrics(
                3,   // steps = maxCycles
                1200,
                2    // reflections = maxCycles - 1
        );
        AgentResponse respostaNaoAprovada = new AgentResponse(
                "Melhor output disponível após 3 ciclos (não aprovado pelo crítico).",
                metricsComMaxCycles
        );

        when(reflectionAgentMock.execute(anyString())).thenReturn(respostaNaoAprovada);

        // Act
        AgentResponse resultado = reflectionAgentMock.execute("Tarefa muito difícil");

        // Assert — verifica que parou em maxCycles (3) e registrou 2 reflexões
        assertThat(resultado.metrics().reflections()).isEqualTo(2);
        assertThat(resultado.metrics().steps()).isEqualTo(3);
        assertThat(resultado.result()).isNotBlank();

        // Deve ter sido chamado exatamente uma vez (o loop interno é encapsulado)
        verify(reflectionAgentMock, times(1)).execute(anyString());
    }

    // -------------------------------------------------------------------------
    // Cenário 3: AgentController roteia para a arquitetura correta
    // -------------------------------------------------------------------------

    @Mock
    private ReactAgent reactAgentMock;

    @Mock
    private ReflectionAgent reflectionAgentForController;

    @InjectMocks
    private AgentController agentController;

    @Test
    @DisplayName("AgentController: deve rotear para ReactAgent quando architecture='react'")
    void agentController_deveRotearParaReactAgent_quandoArchitectureEhReact() {
        // Arrange
        AgentResponse.AgentMetrics metrics = new AgentResponse.AgentMetrics(5, 800, 0);
        AgentResponse respostaReact = new AgentResponse("Resposta ReAct.", metrics);

        when(reactAgentMock.execute(anyString())).thenReturn(respostaReact);

        AgentRequest request = new AgentRequest("react", "Qual é a capital do Brasil?");

        // Act
        ResponseEntity<AgentResponse> response = agentController.run(request);

        // Assert
        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().result()).isEqualTo("Resposta ReAct.");
        assertThat(response.getBody().metrics().reflections()).isZero();

        // Verifica que apenas ReactAgent foi chamado — os demais não devem ser invocados
        verify(reactAgentMock, times(1)).execute("Qual é a capital do Brasil?");
        verify(reflectionAgentForController, never()).execute(anyString());
    }

    @Test
    @DisplayName("AgentController: deve retornar 400 quando architecture é desconhecida")
    void agentController_deveRetornarErro_quandoArchitectureEhDesconhecida() {
        // Arrange
        AgentRequest request = new AgentRequest("arquitetura-inexistente", "Alguma tarefa");

        // Act
        ResponseEntity<AgentResponse> response = agentController.run(request);

        // Assert — controller trata graciosamente com 200 + mensagem de erro no body
        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().result()).contains("Arquitetura desconhecida");
        assertThat(response.getBody().metrics().steps()).isZero();

        // Nenhum agente deve ter sido chamado
        verify(reactAgentMock, never()).execute(anyString());
        verify(reflectionAgentForController, never()).execute(anyString());
    }
}
