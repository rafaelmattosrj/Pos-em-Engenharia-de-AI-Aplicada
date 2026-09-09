package com.opspilot.web;

import com.opspilot.domain.Alert;
import com.opspilot.domain.AlertStatus;
import com.opspilot.domain.ReasoningStrategy;
import com.opspilot.domain.Runbook;
import com.opspilot.domain.Severity;
import com.opspilot.llm.ChatMessage;
import com.opspilot.llm.FakeChatModel;
import com.opspilot.llm.ModelResponse;
import com.opspilot.store.InMemoryOpsStore;
import com.opspilot.store.OpsStore;
import com.opspilot.strategies.ReactStrategy;
import com.opspilot.team.RoleRunners;
import com.opspilot.team.Supervisor;
import com.opspilot.team.TeamRole;
import com.opspilot.team.TeamStrategy;
import com.opspilot.tools.OpsTools;
import com.opspilot.tools.Tool;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.util.List;
import java.util.Map;

/**
 * Porte de src/index.ts (bootstrapOpsPilot + main). Sem OPENROUTER_API_KEY
 * configurada, roda com um FakeChatModel (respostas fixas) -- ver README:
 * uma implementacao real de ChatModel plugaria um provedor OpenAI-compatible
 * aqui no lugar do fake.
 */
public final class ServerMain {

    private ServerMain() {
    }

    public static void main(String[] args) throws IOException {
        OpsStore store = new InMemoryOpsStore();
        seed(store);

        var model = new FakeChatModel(ServerMain::canned);

        List<Tool> analistaTools = List.of(
                OpsTools.listAlerts(store), OpsTools.listIncidents(store),
                OpsTools.consultarRunbook(store), OpsTools.checkProviderStatus());
        List<Tool> executorTools = List.of(
                OpsTools.openIncident(store), OpsTools.resolveIncident(store), OpsTools.listIncidents(store));

        ReasoningStrategy react = new ReactStrategy(model, analistaTools);
        ReasoningStrategy team = new TeamStrategy(
                Supervisor.createDecideNext(model),
                Map.of(
                        TeamRole.ANALISTA, RoleRunners.analista(model, analistaTools),
                        TeamRole.PLANEJADOR, RoleRunners.planejador(model),
                        TeamRole.EXECUTOR, RoleRunners.executor(model, executorTools)),
                1);

        Map<String, ReasoningStrategy> strategies = Map.of(react.name(), react, team.name(), team);

        int port = Integer.parseInt(System.getenv().getOrDefault("PORT", "3000"));
        HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);
        server.createContext("/chat", new ChatHttpHandler(strategies, "team"));
        server.setExecutor(null);
        server.start();
        System.out.println("OpsPilot (nucleo portado) ouvindo em http://localhost:" + port);
    }

    private static void seed(OpsStore store) {
        store.seedAlert(new Alert("alert-1", "checkout-api", "Latencia acima de 2s", Severity.HIGH, AlertStatus.FIRING));
        store.seedRunbook(new Runbook("checkout-api", "1. Verificar pool de conexoes.\n2. Escalar réplicas.\n3. Acionar time de pagamentos se persistir."));
    }

    /** Resposta fixa e determinística -- adequada a demo/smoke test, nao a uso real. */
    private static ModelResponse canned(List<ChatMessage> messages) {
        return ModelResponse.text("{\"next\":\"done\",\"brief\":\"Diagnostico concluido (modelo fake, sem chamada real).\"}");
    }
}
