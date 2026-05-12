# Criação de Agentes Autônomos
**Pós-Graduação em Engenharia de Software com IA Aplicada — UNIPDS**
Rodrigo Fernandes · Rafael Mattos Moreira — Resumo de estudo técnico · 2026

---

## 1. Visão Geral do Módulo

O módulo transforma a perspectiva sobre software: de sistemas que **executam** para sistemas que **decidem**. Um agente autônomo percebe o ambiente, raciocina sobre objetivos, executa ações e avalia resultados em ciclo contínuo — sem intervenção humana a cada passo. O aluno aprende a construir esse ciclo do zero usando Python, definir comportamento via contratos (Markdown + YAML), implementar múltiplas arquiteturas cognitivas (ReAct, Plan-and-Execute, Reflection), integrar com dados reais via Adapters, medir desempenho com Evals e tornar o sistema evolutivo com memória e reflexão.

---

## 2. Conceitos Fundamentais

### Automação vs Agente Autônomo

| Aspecto | Automação Tradicional | Agente Autônomo |
|---|---|---|
| Orientação | Executa **passos** definidos | Persegue **objetivos** |
| Comportamento | Determinístico, fluxo fixo | Adaptativo, decide no contexto |
| Falha | Quebra ou para | Replaneja e tenta alternativas |
| Memória | Stateless | Acumula contexto e aprende |
| Exemplo | Script de CI/CD | Agente de triagem de incidentes |

> 🔑 **Regra fundamental:** Sem objetivo explícito, não existe agente. O que existe é apenas uma automação sofisticada.

### O Agent Loop (Ciclo do Agente)

```
┌─────────────────────────────────────────────────────────┐
│                      AGENT LOOP                         │
│                                                         │
│  PERCEPÇÃO → RACIOCÍNIO → AÇÃO → FEEDBACK → (repeat)   │
│      ↑                                        │         │
│      └────────────────────────────────────────┘         │
└─────────────────────────────────────────────────────────┘
```

| Fase | O que faz | Risco se mal projetada |
|---|---|---|
| **Percepção** | Constrói estado completo (contexto + histórico + sinais de risco) | Decisões baseadas em info incompleta |
| **Raciocínio** | Produz decisão **executável** (não apenas análise) | Loop sem progresso real |
| **Ação** | Executa dentro de capacidades autorizadas | Ações não controladas / destrutivas |
| **Feedback** | Avalia resultado, classifica (sucesso/falha/parcial) | Erros não corrigidos se propagam |

> ⚡ **Performance:** O Agent Loop não é linear. Cada iteração enriquece o estado. A qualidade das decisões melhora não porque o modelo muda, mas porque o contexto evolui.

### Spec Driven Agents

O agente não é definido por código imperativo, mas por **arquivos de especificação** (Markdown + YAML):

| Arquivo | Responsabilidade |
|---|---|
| `agent.md` | Identidade: nome, tipo, objetivo, contrato de saída |
| `loop.md` | Controle do ciclo: max iterações, condições de parada |
| `planner.md` | Estrutura de decisão: formato obrigatório da resposta da LLM |
| `skills.md` | Interfaces das capacidades (o que o agente SABE fazer) |
| `toolbox.md` | Ferramentas autorizadas (o que o agente PODE fazer) |
| `executor.md` | Regras de execução: validação, retry, tratamento de erro |
| `rules.md` | Segurança e limites: rate limits, ações que requerem confirmação |
| `hooks.md` | Pontos de observabilidade: logs, alertas |
| `memory.md` | Gestão de contexto: o que lembrar, o que descartar |

> 💡 **Analogia Java:** Os contratos ↔ `application.yml` + `@Configuration` classes. O runtime ↔ Spring container. A LLM ↔ lógica de negócio injetada. A separação permite reusar o motor para múltiplos agentes.

### Tipos de Agentes

| Tipo | Comportamento | Ideal para |
|---|---|---|
| **Task-Based** | Recebe tarefa, executa, retorna resultado | Problemas bem definidos sem ambiguidade |
| **Interativo** | Faz perguntas antes de agir | Entradas insuficientes ou ambíguas |
| **Goal-Oriented** | Decompõe objetivo em subobjetivos | Problemas complexos, planejamento estruturado |
| **Autônomo** | Responde a eventos com limites rígidos | Monitoramento, alertas, workflows reativos |

> 🔑 O runtime é o **mesmo** para todos os tipos. O que muda é o campo `type` no contrato — como trocar a marcha de um carro.

---

## 3. Detalhes de Implementação

### 3.1 Runtime em Python

O runtime não contém lógica de negócio. Ele interpreta contratos e executa o loop:

```python
class AgentRuntime:
    def __init__(self, agent_path: str):
        self.contracts = ContractLoader.load(agent_path)  # lê todos os .md
        self.state = AgentState.initial()
        self.telemetry = TelemetryCollector()

    def run(self, input: str) -> AgentResult:
        self.state.input = input
        
        for step in range(self.contracts.loop.max_steps):
            # 1. Percepção — monta contexto
            perception = self.build_perception()
            
            # 2. Raciocínio — consulta LLM
            plan = self.planner.plan(perception)  # retorna JSON estruturado
            
            # 3. Ação — executa tool
            result = self.executor.execute(plan.tool, plan.args)
            
            # 4. Feedback — avalia e atualiza estado
            self.state.update(plan, result)
            self.telemetry.record(step, plan, result)
            
            # Critério de parada
            if plan.action == "finish" or self.check_stop_conditions():
                break
        
        return AgentResult(self.state, self.telemetry.generate_trace())
```

> 💡 **Analogia Spring:** `AgentRuntime` ↔ `ApplicationContext`. `ContractLoader` ↔ `@ConfigurationProperties`. `Planner` ↔ `@Service` que chama a LLM.

### 3.2 Arquiteturas Cognitivas

O **planner.md** define qual arquitetura o agente usa. Isso muda como a LLM raciocina:

#### ReAct (Reasoning + Acting)

```
Ciclo: Pensar → Agir → Observar → Repetir
LLM chamada: A CADA iteração
Vantagem: Alta adaptação (descobre caminho conforme avança)
Desvantagem: Alto custo de tokens, menor previsibilidade
```

**Saída estruturada do Planner (ReAct):**
```json
{
  "reasoning": "Recebi alerta de latência. Ainda não sei a causa. Vou consultar métricas do serviço payment-api nas últimas 2h.",
  "action": "call_tool",
  "tool": "get_metrics",
  "args": { "service": "payment-api", "period": "2h" },
  "success_criteria": "Identificar se latência é consistente ou intermitente"
}
```

#### Plan and Execute

```
Ciclo: Planejar tudo → Executar passo a passo
LLM chamada: APENAS 1 VEZ (na geração do plano)
Vantagem: Previsibilidade, menor custo, auditável antes de executar
Desvantagem: Não se adapta a mudanças durante execução
```

**Plano gerado (Plan and Execute):**
```json
{
  "plan": [
    { "step": 1, "tool": "get_deploy_history", "args": { "service": "payment-api", "hours": 24 }, "success_criteria": "Identificar deploys recentes" },
    { "step": 2, "tool": "get_metrics",         "args": { "service": "payment-api", "period": "2h" }, "success_criteria": "Confirmar degradação de latência" },
    { "step": 3, "tool": "get_logs",            "args": { "service": "payment-api", "level": "ERROR" }, "success_criteria": "Encontrar erros correlacionados" },
    { "step": 4, "tool": "save_incident",       "args": { "severity": "high" }, "success_criteria": "Incidente registrado" }
  ],
  "execution_mode": "sequential_no_replan"
}
```

> ⚡ **Performance:** Plan-and-Execute pode reduzir em 70-80% o consumo de tokens vs ReAct para processos bem definidos.

#### Reflection

```
Ciclo: Executar → Avaliar → Criticar → Corrigir (até passar na nota mínima)
LLM chamada: 1 + N reflexões
Vantagem: Alta qualidade de output, erros corrigidos antes de entregar
Desvantagem: Custo e latência adicionais
```

**Critique após execução:**
```json
{
  "score": 0.65,
  "passed": false,
  "feedback": {
    "correctness": "O diagnóstico não considerou o aumento de tráfego paralelo ao deploy",
    "completeness": "Faltou verificar métricas de banco de dados",
    "quality": "A conclusão está correta mas faltam dados específicos"
  }
}
```

#### Comparação das Arquiteturas

| Métrica | ReAct | Plan-Execute | Reflection |
|---|---|---|---|
| Consumo tokens | Alto | Baixo | Alto |
| Previsibilidade | Baixa | Alta | Média |
| Adaptação | Alta | Baixa | Média |
| Qualidade output | Média | Média | Alta |
| Latência | Alta | Baixa | Alta |
| Ideal para | Exploração | Processos estruturados | Alta criticidade |

> 🎯 **Regra de ouro:** A escolha da arquitetura deve ser baseada em **dados** (Evals), não em opinião. Rode benchmarks e compare.

---

### 3.3 Observabilidade — Trace

Logs tradicionais não bastam para agentes. O Trace registra o **porquê** de cada decisão:

```json
{
  "execution_id": "exec-abc123",
  "total_time_ms": 3420,
  "total_tokens": 2847,
  "steps": [
    {
      "step": 1,
      "perception": { "input": "latência alta payment-api", "history": [] },
      "plan": {
        "reasoning": "Ainda não tenho dados. Vou coletar métricas primeiro.",
        "tool": "get_metrics",
        "args": { "service": "payment-api" }
      },
      "result": { "status": "success", "data": { "p99_latency_ms": 2340 } },
      "evaluation": "Latência confirmada alta. Agora buscar histórico de deploy."
    }
  ],
  "health_metrics": {
    "tool_success_rate": 0.92,
    "circuit_breaker_activations": 0,
    "payload_validation_errors": 1
  }
}
```

> 💡 **Analogia Java:** Trace ↔ Distributed Tracing (Jaeger/Zipkin) mas para decisões de LLM, não spans de microserviços.

Os **quatro níveis de observabilidade:**

| Nível | Ferramenta | Propósito |
|---|---|---|
| Hooks | Eventos em tempo real | Alertas imediatos |
| Dashboard | Visualização live | Monitoramento operacional |
| Trace | Registro completo | Debug e auditoria |
| Análise automatizada | Agente analisador | Diagnóstico e recomendações |

---

### 3.4 Evals e Benchmarks

```python
# Estrutura de um dataset de eval
eval_dataset = [
  {
    "id": "deploy-latency-001",
    "input": "Latência alta em payment-api após deploy das 14h",
    "difficulty": "medium",
    "expected_tools": ["get_deploy_history", "get_metrics", "get_logs"]
  }
]

# Métricas do benchmark
metrics = {
  "completion_rate": 0.87,      # % cenários concluídos
  "avg_steps": 4.2,             # média de etapas
  "avg_tokens": 3150,           # média de tokens
  "tool_success_rate": 0.91,    # % tools executadas com sucesso
  "tool_coverage": 0.78         # % tools esperadas efetivamente usadas
}
```

> 🔑 **Eval não é teste unitário.** Testes validam código. Evals validam **comportamento** do agente em diferentes cenários.

---

### 3.5 Adapters — Integração com Dados Reais

Dados mockados levam a diagnósticos incorretos e falsos positivos. A solução é o padrão Adapter:

```python
# Contrato da skill (não muda)
skill: get_metrics
  description: "Busca métricas de latência do serviço"
  input: { service: str, period: str }
  output: { p50: float, p99: float, error_rate: float }
  
# Implementação pode mudar sem alterar o contrato
  implementation:
    type: rest  # antes era: mock
    url: "http://monitoring.internal/metrics/{service}"
    timeout_ms: 5000
```

| Tipo de Adapter | Uso | Cuidado |
|---|---|---|
| `mock` | Desenvolvimento e testes | Dados fictícios = diagnósticos irreais |
| `rest` | APIs externas | Rate limiting, timeout, autenticação |
| `database` | Consultas diretas | Apenas leitura, queries parametrizadas, SQL injection |
| `mcp` | Servidores MCP | Segurança do servidor MCP |

> 🔑 **Segurança Database Adapter:** Sempre `READ ONLY` + queries parametrizadas. Nunca permitir que a LLM construa SQL diretamente — risco de SQL Injection.

---

### 3.6 Tool Selection Eval

Além de avaliar o agente como todo, é preciso avaliar **decisões individuais**:

| Métrica | Pergunta | Limite aceitável |
|---|---|---|
| `tool_selection_accuracy` | Escolheu a tool correta? | ≥ 80% |
| `argument_accuracy` | Argumentos corretos? | ≥ 85% |
| `unnecessary_calls_rate` | Chamou tools desnecessárias? | ≤ 10% |
| `wrong_tool_rate` | Usou tool errada? | ≤ 15% |

**Causa mais comum de erros:** Ambiguidade nas descrições das skills, não o modelo.

```
# Antes (72% acurácia):
skill: get_logs — "Busca logs do sistema"

# Depois (91% acurácia):
skill: get_recent_logs  — "Busca logs via API REST das últimas 60 min para diagnóstico imediato"
skill: get_historic_logs — "Busca logs históricos via banco de dados para análise temporal comparativa"
```

> ✅ **De 72% para 91% de acurácia sem mudar modelo, prompt principal ou código** — apenas reescrevendo as descrições.

---

### 3.7 Memória — Os 4 Tipos

| Tipo | Duração | O que armazena | Analogia humana |
|---|---|---|---|
| **Curta** | Execução atual | Estado em progresso, decisões do ciclo | Mesa de trabalho |
| **Longa** | Persistente | Fatos confirmados, configurações do sistema | Caderno de referência |
| **Episódica** | Persistente | Histórico de execuções, problemas resolvidos | Diário profissional |
| **Contextual** | Persistente + busca semântica | Recuperação por similaridade (embeddings) | Motor de busca interno |

**Agent Loop com memória:**
```
ANTES: Percepção → Planejamento → Ação → Avaliação
DEPOIS: Percepção → Recuperar contexto → Planejamento → Ação → Avaliação → Persistir memória
```

**Impacto prático:** Em cenário de monitoramento, redução de ~37% nas chamadas após acumular histórico relevante.

> 🔑 **Regra crítica:** Nunca armazenar dados sensíveis (tokens, senhas, credenciais) na memória. Armazenar apenas fatos **confirmados por ferramentas**.

---

### 3.8 Embeddings e Reflexão Evolutiva

**Por que busca por texto não basta:**
```
"Latência alta no checkout"  ≠  "Tempo de resposta degradado"  (busca textual)
"Latência alta no checkout"  ≈  "Tempo de resposta degradado"  (busca semântica)
```

**Embed Adapter:**
```python
class EmbedAdapter:
    def index(self, text: str, metadata: dict) -> None:
        embedding = embedding_model.encode(text)
        vector_store.add(embedding, metadata)
    
    def search(self, query: str, threshold: float = 0.7) -> list[dict]:
        query_embedding = embedding_model.encode(query)
        results = vector_store.similarity_search(query_embedding)
        return [r for r in results if r.score >= threshold]
```

**Reflexão Evolutiva:** O agente analisa execuções passadas e extrai lições:

```json
{
  "lesson": "Em falhas de latência do payment-api após deploy, verificar o histórico de deploys ANTES das métricas reduz o número de etapas de 5 para 3",
  "situation": "latência alta após deploy",
  "action": "get_deploy_history primeiro",
  "result": "sucesso",
  "generalizability": "alta"
}
```

> ⚡ Diferença fundamental: **Memória armazena o passado. Reflexão transforma o passado em diretriz para o futuro.**

---

### 3.9 Evals de Memória

A memória pode melhorar, ser irrelevante ou **prejudicar** o agente. Sempre medir:

| Métrica | O que mede | Ação se baixa |
|---|---|---|
| `retrieval_precision` | % fragmentos relevantes recuperados | Aumentar threshold de similaridade |
| `retrieval_recall` | % informações importantes recuperadas | Ajustar critérios de busca |
| `memory_utilization` | % contexto recuperado efetivamente usado | Revisar regras do Planner |
| `hallucination_from_memory` | Informações inventadas baseadas na memória | Implementar políticas de expiração |
| **`decision_improvement`** | **Comparação com/sem memória** | **Métrica central** |
| `lesson_quality` | Qualidade das lições extraídas | Ajustar processo de reflexão |

> 🔑 Um `decision_improvement` negativo significa que a memória está **prejudicando** o agente. Isso acontece com dados desatualizados — implemente políticas de expiração.

---

## 4. Analogias Java/Spring Boot

| Conceito Agente | Equivalente Java/Spring Boot |
|---|---|
| Agent Loop | `@Scheduled` task com lógica de negócio iterativa |
| Contratos (`.md`) | `application.yml` + `@ConfigurationProperties` |
| Runtime | Spring Application Context |
| Planner | `@Service` que chama LLM API |
| Executor | `@Component` que despacha para ferramentas |
| Skills | Interfaces de Service (sem implementação) |
| Toolbox | Lista autorizada de `@Bean` disponíveis |
| Rules | `@PreAuthorize` + validações de entrada |
| Hooks | `@EventListener` / AOP `@Around` |
| Memory Adapter | `@Repository` com múltiplas implementações |
| Embed Adapter | Spring Data com pgvector/Elasticsearch |
| Trace | OpenTelemetry / Micrometer distributed tracing |
| Eval | Suite de testes de comportamento (não unitários) |
| Circuit Breaker | Resilience4j `@CircuitBreaker` |
| Reflection | Análise pós-execução com feedback loop |

> 💡 **Analogia arquitetural completa:**
> ```
> LLM = motor de decisão (não o agente!)
> Contratos = especificação do comportamento
> Runtime = executor genérico (como Spring container)
> Skills/Toolbox = capacidades disponíveis
> Rules = guardrails de segurança
> Memory = estado persistente entre execuções
> Trace = observabilidade das decisões
> ```

---

## 5. Escalabilidade e Produção

### Quando Agentes São Adequados

```
✅ USE AGENTES QUANDO:
  - O problema requer julgamento contextual
  - O caminho de solução não é totalmente previsível
  - A tarefa combina múltiplas ferramentas de forma flexível
  - Há valor em aprender com execuções anteriores

❌ NÃO USE AGENTES QUANDO:
  - O fluxo é 100% determinístico e bem definido
  - A latência é crítica (agentes são mais lentos)
  - O custo de tokens é proibitivo para o volume
  - O domínio é muito restrito (automação simples basta)
```

### Proteções Obrigatórias em Produção

```yaml
# rules.md — contratos de segurança
limits:
  max_steps: 10
  max_tokens: 50000
  max_tool_calls_per_tool: 3
  max_execution_time_seconds: 60

safety:
  require_confirmation:
    - tool: rollback_deploy
    - tool: send_notification
  never_allow:
    - tool: delete_database
    - action: store_credentials_in_memory

circuit_breaker:
  activate_on: consecutive_failures >= 3
  recovery_wait_seconds: 30
```

### Antipadrões a Evitar

| Antipadrão | Consequência | Solução |
|---|---|---|
| Loop sem critério de parada | Consumo infinito de recursos | `max_steps` obrigatório no `loop.md` |
| Ações sem controle | Operações destrutivas irreversíveis | `toolbox.md` lista explícita |
| Mock em produção | Diagnósticos falsos | Adapters REST/DB reais |
| Sem observabilidade | Debug impossível | Hooks + Trace obrigatórios |
| Skills com descrições ambíguas | Baixa acurácia de seleção | Tool Selection Eval contínuo |
| Memória sem expiração | Dados desatualizados prejudicam decisões | Políticas de TTL na memória longa |

---

## 6. Conexões entre Módulos

| Módulo | Como se conecta com Agentes |
|---|---|
| **Módulo 01** (Fundamentos LLMs) | LLM é o "motor de raciocínio" do agente — entender tokens, temperatura e contexto é base |
| **Módulo 02** (APIs LLMs) | O Planner chama APIs de LLMs; RAG alimenta o contexto do agente |
| **Módulo 03** (MCP) | Agentes consomem MCPs como ferramentas via MCP Adapter |
| **Módulo 04** (Agentes — este) | Base para construir sistemas multi-agente no futuro |

### Idéias de Projetos de Portfólio

| Projeto | Tipo de Agente | Skills |
|---|---|---|
| Triagem de incidentes | Task-Based | get_metrics, get_logs, create_incident |
| Revisão de pull requests | Goal-Oriented | read_diff, check_patterns, add_comment |
| Decomposição de backlog | Goal-Oriented | parse_requirements, create_epics, estimate_effort |
| Validação de dados | Task-Based | query_database, check_integrity, generate_report |
| Onboarding de devs | Interativo | read_docs, search_codebase, answer_questions |

### Glossário do Módulo

| Termo | Definição |
|---|---|
| **Agent Loop** | Ciclo contínuo: Percepção → Raciocínio → Ação → Feedback |
| **Spec Driven Agents** | Comportamento definido por arquivos de especificação, não por código |
| **Planner** | Componente que consulta a LLM e retorna decisão estruturada |
| **Executor** | Componente que executa a ferramenta escolhida pelo Planner |
| **ReAct** | Arquitetura cognitiva: raciocínio incremental a cada etapa |
| **Plan-and-Execute** | Arquitetura cognitiva: plano completo antes de executar |
| **Reflection** | Arquitetura cognitiva: autoavaliação antes de finalizar |
| **Trace** | Registro completo da execução incluindo raciocínios e decisões |
| **Eval** | Avaliação automatizada do comportamento do agente com dataset + métricas |
| **Adapter** | Camada de conexão com fontes de dados reais (REST, DB, MCP) |
| **Tool Selection Accuracy** | % de vezes que o agente escolheu a tool correta no contexto certo |
| **Decision Improvement** | Comparação de desempenho com e sem memória ativa |
| **Embedding** | Representação numérica de texto que captura significado semântico |
| **Reflexão Evolutiva** | Processo de extrair lições de execuções passadas e aplicar em futuras |
| **Circuit Breaker** | Proteção que interrompe execução em caso de falhas consecutivas |
