# Criação de Agentes Autônomos — Explicação Detalhada
**Pós-Graduação em Engenharia de Software com IA Aplicada — UNIPDS**
Rodrigo Fernandes · Rafael Mattos Moreira · 2026

---

## Módulo 1: Introdução aos Agentes Autônomos

### Capítulo 1: Arquitetura de Agents

#### O Problema dos Sistemas Tradicionais

Sistemas tradicionais são construídos para seguir um roteiro. São baseados em lógica determinística, em decisões binárias, em `if` e `else`. O problema é que o mundo real não funciona dessa forma.

Quando um fluxo não previsto surge, o sistema simplesmente para de responder corretamente — não necessariamente quebra, não lança exceção, não gera logs úteis. Ele apenas deixa de funcionar como deveria.

#### Execução vs Decisão

> 🔑 Grande parte dos sistemas que construímos hoje são orientados à **execução**. Mas muitas atividades do dia a dia envolvem **julgamento, contexto e tomada de decisão**.

Exemplos de tarefas que exigem decisão (não apenas execução):
- Revisão de pull requests: equilibrar risco e qualidade
- Triagem de incidentes: decidir o que investigar primeiro
- Avaliação de mudanças arquiteturais e seus impactos

#### O Limite das Automações

Automações são excelentes para tarefas bem definidas com regras claras. Mas possuem uma **limitação estrutural: não tomam decisões**. Elas apenas seguem instruções.

Mesmo quando utilizamos LLMs de forma reativa (prompt → resposta), estamos operando nesse modelo. O modelo responde ao que foi solicitado, mas não toma iniciativa, não define ações e não evolui o contexto por conta própria.

#### O Que é um Agente Autônomo

Um agente autônomo é um sistema que:

1. **Percebe** o ambiente
2. **Toma decisões** com base em objetivos
3. **Executa ações** de forma autônoma
4. **Avalia** resultados e repete

> 🔑 O mais importante: o agente opera em **loop**. Ele não executa uma única vez e encerra. Ele continua avaliando o estado e ajustando comportamento conforme necessário.

#### A Importância do Objetivo

**Sem objetivo explícito, não existe agente. O que existe é apenas uma automação.**

- Uma automação executa passos.
- Um agente persegue um objetivo.

O objetivo é o que orienta **todas** as decisões do agente. É o critério que permite avaliar se uma ação foi bem-sucedida ou não.

#### Spec Driven Agents

O comportamento do agente é definido por **especificações escritas em Markdown**:

```
agents/monitor-agent/
  ├── agent.md      ← identidade e objetivo
  ├── loop.md       ← controle do ciclo
  ├── planner.md    ← estrutura de decisão
  ├── skills.md     ← capacidades disponíveis
  ├── toolbox.md    ← ferramentas autorizadas
  ├── executor.md   ← regras de execução
  ├── rules.md      ← segurança e limites
  ├── hooks.md      ← observabilidade
  └── memory.md     ← gestão de contexto
```

Isso transforma o desenvolvimento de agentes em um **processo estruturado**, não tentativa e erro.

---

### Capítulo 2: Agent Loop

#### As Quatro Fases

**Percepção — Construção de Estado**

Percepção não é apenas leitura de input. É a construção de um **estado completo**. Um alerta de latência isolado não é suficiente para decisão. É necessário enriquecer com:

- Qual serviço está sendo afetado e desde quando
- Se houve deploy recente
- Se esse comportamento já ocorreu anteriormente
- Sinais de risco (tentativas repetidas, falta de informação, crescimento de custo)

> 🔑 Sem contexto correto, não existe decisão correta.

**Raciocínio — Da Análise à Decisão Executável**

Raciocinar não é apenas pensar bem. É **produzir uma decisão executável**.

❌ Análise (não executa): "O problema pode estar relacionado ao banco de dados, à rede ou ao cache."

✅ Decisão executável: "Consultar métricas de latência do payment-api na última hora para verificar degradação após deploy."

Toda decisão precisa estar associada a um **critério de sucesso**. Sem esse critério, o agente não consegue avaliar se deve continuar, replanejar ou encerrar.

**Ação — Execução Controlada**

Ação não significa liberdade irrestrita. O agente deve operar **estritamente dentro das capacidades autorizadas**:

- **Limitada** — toolbox define o que pode ser usado
- **Autorizada** — rules define o que requer confirmação
- **Observável** — hooks registram cada execução

**Feedback — Avaliação e Aprendizado**

Após executar, o resultado é classificado:
- `success` — continua ou encerra
- `partial` — pode replanejar
- `failure` — tenta alternativas ou encerra com erro

Sem atualizar o estado, o agente não aprende e **tende a repetir os mesmos erros indefinidamente**.

#### O Problema dos Loops Descontrolados

Um agente mal projetado pode continuar executando indefinidamente. Diferente de automações que falham rapidamente, um agente em loop infinito:
- Não apenas continua executando
- Ele continua **decidindo** — e potencialmente causando dano

> ⚡ O critério de parada não é opcional. É responsabilidade essencial da engenharia.

**Exemplo completo do loop:**

```
1. Percepção: identifica latência alta em payment-api após alerta
2. Raciocínio: decide consultar métricas (ainda sem dados)
3. Ação: executa get_metrics(service="payment-api", period="1h")
4. Feedback: p99=2340ms, muito alto. Próximo passo: verificar deploys

Nova iteração:
1. Percepção: atualiza estado com dados de métricas
2. Raciocínio: vê que métricas pioraram após 14h — suspeita de deploy
3. Ação: executa get_deploy_history(service="payment-api", since="13h")
4. Feedback: deploy confirmado às 14h02. Causa identificada.

Nova iteração:
1. Planejamento: registrar incidente
2. Ação: save_incident(severity="high", cause="deploy 14h02")
3. Feedback: success → critério de parada atingido → encerra
```

---

### Capítulo 3: Contratos

#### O Papel da LLM

> 🔑 A LLM não é o agente. Ela é o **motor de raciocínio**. Quem decide é o agente — definido pelos contratos.

A LLM não define:
- Limites de execução
- Estrutura do sistema
- Ferramentas disponíveis
- Regras de segurança

Ela opera **dentro** de regras previamente estabelecidas pelos contratos.

#### agent.md — Identidade

```yaml
agent:
  name: "MonitorAgent"
  description: "Agente de monitoramento de produção para triagem de incidentes"
  type: "task-based"
  
  objective: |
    Analisar alertas de produção, identificar causas raiz e gerar relatório
    de diagnóstico com evidências coletadas, classificação de severidade e
    próximas ações recomendadas.
    
  output_contract:
    format: "json"
    required_fields:
      - diagnosis
      - severity
      - evidence
      - recommended_actions
```

#### loop.md — Controle do Ciclo

```yaml
loop:
  objective: "Identificar causa raiz do alerta e gerar diagnóstico completo"
  max_steps: 10
  max_tokens: 50000
  max_execution_time_seconds: 120
  
  stop_conditions:
    - condition: "action == 'finish'"
      reason: "Objetivo alcançado"
    - condition: "steps >= max_steps"
      reason: "Limite de etapas excedido"
    - condition: "no_progress_for_steps >= 3"
      reason: "Sem progresso detectado"
    - condition: "tokens >= max_tokens"
      reason: "Limite de contexto atingido"
```

#### planner.md — Estrutura de Decisão

O planner define o **contrato da resposta da LLM**. Não é um prompt livre — é um contrato:

```yaml
planner:
  architecture: "react"  # react | plan-and-execute | reflection
  
  output_schema:
    reasoning: "string - O que o agente já sabe e por que escolheu esta ação"
    action: "enum: call_tool | finish | request_info"
    tool: "string - Nome da tool (se action == call_tool)"
    args: "object - Argumentos da tool"
    success_criteria: "string - O que constitui sucesso desta etapa"
    
  rules:
    - "SEMPRE registrar evidências antes de finalizar"
    - "NUNCA assumir causa raiz sem ao menos 2 ferramentas de confirmação"
    - "SE dados insuficientes → usar request_info, NÃO inventar"
```

#### skills.md vs toolbox.md — Saber vs Poder

**Skills definem a interface** (o que o agente sabe fazer):

```yaml
skills:
  - name: get_metrics
    description: "Busca métricas de latência e erros do serviço nas últimas N horas"
    input:
      service: "string - Nome do serviço"
      period: "string - Ex: '1h', '24h'"
    output:
      p50_ms: "float"
      p99_ms: "float"
      error_rate: "float"
```

**Toolbox autoriza o uso** (o que pode ser executado):

```yaml
toolbox:
  allowed:
    - get_metrics
    - get_logs
    - get_deploy_history
    - save_incident
  # get_database_password NÃO está aqui → agente não pode usar
```

> 🔑 Se uma skill está definida mas não está na toolbox, o agente **não pode** utilizá-la. Isso cria dupla camada de segurança.

#### rules.md — Segurança

```yaml
rules:
  limits:
    max_calls_per_tool: 3
    
  require_human_confirmation:
    - tool: rollback_deploy
      reason: "Operação destrutiva irreversível"
    - tool: send_critical_alert
      condition: "severity == 'critical'"
      
  mandatory_behaviors:
    - "DEVE salvar evidências antes de chamar 'finish'"
    - "NUNCA armazenar tokens ou senhas na memória"
```

---

### Capítulo 4: Runtime

#### Estrutura do Runtime (Python)

```python
class AgentRuntime:
    """Motor genérico — não sabe o que o agente faz, só como executar."""
    
    def __init__(self, agent_path: str):
        # Carrega YAML de todos os .md do agente
        self.contracts = ContractLoader.load(agent_path)
        self.state = AgentState.initial()
        self.planner = Planner(self.contracts.planner)
        self.executor = Executor(self.contracts.executor, self.contracts.toolbox)
        self.telemetry = TelemetryCollector()
    
    def run(self, input: str) -> AgentResult:
        self.state.input = input
        
        for step in range(self.contracts.loop.max_steps):
            if self._check_stop_conditions(): break
            
            # 1. Percepção
            perception = self._build_perception()
            
            # 2. Raciocínio (chama LLM)
            plan = self.planner.plan(perception)
            
            # Circuit Breaker — protege contra respostas inválidas
            if not self._validate_plan(plan):
                self.state.circuit_breaker_count += 1
                continue
            
            # 3. Ação
            result = self.executor.execute(plan.tool, plan.args)
            
            # 4. Feedback
            self.state.update(plan, result)
            self.telemetry.record(step, plan, result)
            
            if plan.action == "finish": break
        
        return AgentResult(self.state, self.telemetry.generate_trace())
```

#### O Planejador e a LLM

```python
class Planner:
    def plan(self, perception: Perception) -> Plan:
        # Monta o prompt com contexto completo
        prompt = self._build_prompt(perception)
        
        # Chama LLM — retorna JSON estruturado (não texto livre)
        response = self.llm.chat(prompt, response_format=self.contracts.output_schema)
        
        return Plan(**response)
    
    def _build_prompt(self, p: Perception) -> str:
        return f"""
Objective: {self.contracts.objective}
Current State: {p.state_summary}
History: {p.decision_history}
Available Tools: {p.available_tools}
Memory Context: {p.retrieved_context}
Lessons Learned: {p.relevant_lessons}

RULES:
{self.contracts.rules}

Respond ONLY with valid JSON matching the schema:
{self.contracts.output_schema}
"""
```

#### Telemetria e Trace

```python
# Trace gerado automaticamente após cada execução
{
  "execution_id": "exec-abc123",
  "agent": "MonitorAgent",
  "total_time_ms": 3420,
  "total_tokens": 2847,
  "steps": [
    {
      "step": 1,
      "perception_tokens": 312,
      "plan": {
        "reasoning": "Ainda sem dados. Preciso coletar métricas primeiro.",
        "tool": "get_metrics",
        "args": { "service": "payment-api", "period": "2h" }
      },
      "result": { "status": "success", "latency_ms": 245 },
      "evaluation": "Latência p99 = 2340ms. Alta. Próximo: verificar deploy."
    }
  ],
  "health_metrics": {
    "tool_success_rate": 0.92,
    "circuit_breaker_activations": 1
  }
}
```

---

### Capítulos 5 e 6: Observabilidade e Tipos de Agentes

#### Os 4 Níveis de Observabilidade

**Nível 1 — Hooks (tempo real):**
```yaml
hooks:
  on_step_start: log_step_beginning
  on_tool_call: log_tool_invocation  
  on_error: alert_on_call + log_error
  on_finish: save_trace + notify_completion
```

**Nível 2 — Dashboard:** Visualização live de progresso, tokens, qualidade.

**Nível 3 — Trace:** JSON completo com cada decisão, reasoning, resultado e avaliação.

**Nível 4 — Agente Analisador:** Um agente que analisa o trace de outro e gera diagnóstico:

```
Input: trace.json de uma execução
Output:
  - Diagnóstico de comportamento
  - Identificação de gargalos (onde foi mais lento)
  - Violações de regras
  - Recomendações de melhoria dos contratos
```

#### Tipos de Agentes na Prática

**Backlog Decomposer (Goal-Oriented):**
```
Input: "Sistema de pagamentos precisa suportar Pix instantâneo"
Output:
  Épico 1: Integração com Banco Central (Pix API)
    História 1.1: Cadastrar chave Pix por CPF/CNPJ [8 pts]
    História 1.2: Realizar transferência instantânea [13 pts]
    CTs: deve processar em < 3s; deve rejeitar valores inválidos
  
  Riscos:
    - Disponibilidade da API BC pode impactar SLA
  
  Perguntas:
    - Qual o volume esperado de transações/hora?
```

---

## Módulo 2: Raciocínio e Tomada de Decisão

### Capítulo 1: Arquiteturas Cognitivas

#### O Problema Central

Dois agentes com o mesmo contexto, as mesmas ferramentas e o mesmo objetivo podem produzir resultados completamente diferentes. A diferença está na **forma como o raciocínio é estruturado**.

Isso é o que define a arquitetura cognitiva.

#### ReAct com Reasoning Trace Explícito

Antes (reasoning implícito — dificulta debug):
```
Ação executada: get_metrics
Resultado: p99 = 2340ms
```

Depois (reasoning explícito — trace completo):
```json
{
  "reasoning": "Recebi alerta de latência. Não tenho dados ainda. Preciso coletar métricas do payment-api nas últimas 2h para confirmar a magnitude do problema e identificar se é consistente ou intermitente.",
  "action": "call_tool",
  "tool": "get_metrics",
  "args": { "service": "payment-api", "period": "2h" },
  "success_criteria": "Obter valores de p50, p99 e error_rate para avaliar gravidade"
}
```

> 🔑 Quando o agente é obrigado a explicar seu raciocínio, a qualidade das decisões melhora — o modelo evita escolhas superficiais porque precisa justificá-las.

#### Risco de Loop no ReAct

```
Problema: Loop exploratório sem critério de parada
  → consulta métricas (não conclusivo)
  → consulta novamente com período diferente (não conclusivo)
  → consulta novamente com serviço diferente
  → ... indefinidamente

Solução: limites no loop.md + detecção de falta de progresso
```

---

### Capítulo 2: Plan-and-Execute e Reflection

#### Plan-and-Execute — Quando Usar

Plan-and-Execute funciona bem quando:
- O processo é **conhecido e estável**
- As etapas são **previsíveis**
- Há necessidade de **auditabilidade** (revisar o plano antes de executar)
- O custo de tokens é uma restrição

**Falha silenciosa:** Se um passo retorna resultado inesperado, o agente continua executando com dados incorretos. Use apenas quando o ambiente é estável.

**Exemplo — Plano completo gerado antes de executar:**
```
[14:32:01] Gerado plano com 4 etapas:
  Step 1: get_deploy_history(service="payment-api", hours=24)
  Step 2: get_metrics(service="payment-api", period="2h")
  Step 3: get_logs(service="payment-api", level="ERROR")  
  Step 4: save_incident(severity="high")

[14:32:01] Iniciando execução sequencial...
[14:32:02] Step 1: ✅ deploy encontrado às 14h02
[14:32:03] Step 2: ✅ p99 = 2340ms confirmado
[14:32:04] Step 3: ✅ 847 erros correlacionados
[14:32:05] Step 4: ✅ incidente criado #INC-4521

Tokens consumidos: 1 chamada LLM (387 tokens) vs ReAct: 8 chamadas (3.200 tokens)
```

#### Reflection — Autoavaliação

**Contrato do Critique:**
```yaml
critique:
  evaluate_dimensions:
    correctness:
      question: "O diagnóstico está correto e baseado em evidências?"
      weight: 0.4
    completeness:
      question: "Todas as evidências disponíveis foram consideradas?"
      weight: 0.3
    quality:
      question: "Os dados são específicos e acionáveis?"
      weight: 0.3
  
  min_score_to_accept: 0.75
  max_reflection_cycles: 3
```

**Ciclo de Reflexão:**
```
Execução → Score: 0.65 (abaixo de 0.75)
  Feedback: "Faltou verificar métricas de banco de dados"
  
Re-execução com contexto atualizado → Score: 0.82 (acima de 0.75)
  ✅ Aceito
```

---

### Capítulo 3: Evals e Frameworks de Mercado

#### Benchmark Comparativo

**Dataset:** Mesmos cenários aplicados a todas as arquiteturas.

```python
# Resultado real de benchmark com 10 cenários
results = {
  "react": {
    "completion_rate": 0.90,
    "avg_tokens": 3150,
    "avg_steps": 4.8,
    "tool_coverage": 0.85,
    "verdict": "PASSED"
  },
  "plan_execute": {
    "completion_rate": 0.80,
    "avg_tokens": 870,   # muito menor!
    "avg_steps": 4.0,
    "tool_coverage": 0.72,
    "verdict": "PASSED"
  },
  "reflection": {
    "completion_rate": 0.90,
    "avg_tokens": 4200,
    "avg_steps": 5.1,
    "reflection_avg": 1.3,
    "verdict": "PASSED"
  }
}
```

**Interpretação:**
- Plan-Execute: 3,6x mais barato que ReAct em tokens, mas menor cobertura de tools
- ReAct: maior cobertura e adaptação, maior custo
- Reflection: melhor qualidade quando há erro, mas 33% mais caro que ReAct

> 🎯 Não existe "melhor" — existe **melhor para o problema**. Escolha baseada em dados.

#### Equivalência com Frameworks de Mercado

| Este Módulo | LangChain | LangGraph |
|---|---|---|
| Skills | Tools | Tools |
| Planner | Agent (LCEL Chain) | Node + LLM |
| Executor | AgentExecutor | Edge + Tool Node |
| Loop | AgentExecutor loop | StateGraph cycle |
| Rules | Guardrails | Conditional edges |
| Observabilidade | Callbacks | LangSmith tracing |

> 🔑 Frameworks mudam constantemente. Conceitos permanecem. Quem entende o conceito se torna independente de framework.

---

## Módulo 3: Integração com o Mundo Real

### Capítulo 1: De Mock para Real

#### Os Três Tipos de Problema com Mock

```python
# Falso positivo: mock gera latência alta quando sistema está saudável
def get_metrics_mock(service, period):
    return {
        "p99_ms": random.randint(800, 3000),  # aleatório!
        "error_rate": random.uniform(0, 0.1)
    }

# Problema: agente sempre encontra "problemas" mesmo quando não existem
```

```python
# Solução: REST Adapter com dados reais e consistentes
class MetricsRestAdapter(BaseAdapter):
    def execute(self, tool: str, args: dict) -> dict:
        response = requests.get(
            f"{self.config.base_url}/metrics/{args['service']}",
            params={"period": args.get("period", "1h")},
            headers={"Authorization": f"Bearer {self.config.token}"},
            timeout=self.config.timeout_ms / 1000
        )
        response.raise_for_status()
        return response.json()
```

#### Tipos de Adapter

```yaml
# skills.md — implementação configurável por tipo
skills:
  - name: get_metrics
    implementation:
      type: rest
      base_url: "http://monitoring.internal"
      endpoint: "/metrics/{service}"
      timeout_ms: 5000
      retry:
        max_attempts: 3
        backoff_ms: 1000
```

```python
# resolver_adapter.py — seleciona implementação dinamicamente
class ResolverAdapter:
    ADAPTERS = {
        "mock": MockAdapter,
        "rest": RestAdapter,
        "database": DatabaseAdapter,
        "mcp": McpAdapter
    }
    
    def resolve(self, skill: Skill) -> BaseAdapter:
        adapter_type = skill.implementation.type
        if adapter_type not in self.ADAPTERS:
            return MockAdapter(skill)  # fallback seguro
        return self.ADAPTERS[adapter_type](skill)
```

---

### Capítulo 2: Database e MCP Adapter

#### Database Adapter — Segurança Obrigatória

```python
class DatabaseAdapter(BaseAdapter):
    def execute(self, tool: str, args: dict) -> dict:
        query_config = self.skill.implementation.query
        
        # Queries parametrizadas — NUNCA construção dinâmica de SQL
        # Protege contra SQL Injection
        with self.db.read_only_connection() as conn:  # read-only sempre
            result = conn.execute(
                query_config.sql,          # SQL fixo no contrato
                self._sanitize(args)        # parâmetros separados
            )
            
            # Limite de registros — evita sobrecarga
            rows = result.fetchmany(query_config.max_rows or 100)
            return {"rows": rows, "count": len(rows)}
    
    def _sanitize(self, args: dict) -> dict:
        """Remove campos não esperados pelo contrato."""
        allowed = set(self.skill.input_schema.keys())
        return {k: v for k, v in args.items() if k in allowed}
```

> 🔑 **Três regras inegociáveis para Database Adapter:**
> 1. Apenas leitura (`READ ONLY`)
> 2. Queries parametrizadas (não SQL dinâmico)
> 3. Validação rigorosa de entrada

**Configuração no contrato:**
```yaml
# skills.md
- name: query_incidents
  implementation:
    type: database
    query:
      sql: "SELECT * FROM incidents WHERE service = :service AND created_at > NOW() - INTERVAL :hours HOUR"
      max_rows: 50
      timeout_seconds: 10
```

#### MCP Adapter

```python
class McpAdapter(BaseAdapter):
    def execute(self, tool: str, args: dict) -> dict:
        # Conecta ao servidor MCP (STDIO ou HTTP)
        with McpClient(self.skill.implementation.server_config) as client:
            result = client.call_tool(tool, args)
            return result.content
```

> 💡 **Por que MCP Adapter?** Permite que o agente use ferramentas MCP externas (MongoDB, File System, qualquer servidor MCP publicado) sem código específico de integração.

---

### Capítulo 3: Tool Selection Eval

#### Por que Avaliar Decisões Individuais

Um agente pode completar todas as etapas e finalizar com sucesso — e ainda assim ter tomado **decisões ruins**. A diferença:

```
Cenário: Alerta de latência após deploy às 14h

Comportamento A (correto): get_deploy_history primeiro → contexto relevante
Comportamento B (incorreto): get_metrics primeiro → perde o contexto do deploy

Ambos finalizam com sucesso.
Apenas A tomou a decisão certa para o contexto.
```

#### Dataset por Etapa do Loop

```python
tool_selection_dataset = [
  {
    "id": "deploy-latency-step1",
    "context": "Alerta: latência alta em payment-api. Deploy ocorreu há 30min.",
    "loop_step": 1,
    "history": [],
    "expected_tool": "get_deploy_history",
    "expected_args_keywords": ["payment-api"],
    "forbidden_tools": ["save_incident"],  # muito cedo para salvar
    "rationale": "Contexto menciona deploy explicitamente — deve ser investigado primeiro"
  }
]
```

#### O Impacto da Temperatura na Eval

```python
# LLMs têm aleatoriedade — resultados inconsistentes entre execuções

# Solução: temperatura 0 + seed fixo para evals reproduzíveis
llm_config = {
  "temperature": 0,    # sem aleatoriedade
  "seed": 42           # determinismo garantido
}
```

> ⚡ Sem controle de temperatura, evals são inúteis — resultados diferentes a cada run.

#### Diagnóstico por Métrica

| Métrica baixa | Causa provável | Solução |
|---|---|---|
| Tool Selection < 80% | Skills com descrições ambíguas | Reescrever descriptions mais específicas |
| Argument Accuracy < 85% | Contexto insuficiente no percepção | Enriquecer construção do estado |
| Unnecessary Calls > 10% | Planner sem regras de contenção | Adicionar regra "não repita tools sem novos dados" |
| Wrong Tool Rate > 15% | Tools com domínios sobrepostos | Separar claramente quando usar cada uma |

---

## Módulo 4: Memória e Evolução

### Capítulo 1: O Agente que Lembra

#### Impacto Real da Falta de Memória

```
Execução 1 (segunda-feira):
  → 5 etapas para diagnosticar: latência alta no payment-api causada por índice faltante no banco

Execução 2 (quarta-feira — mesmo problema):
  → Sem memória: 5 etapas novamente, reprocessamento completo
  → Com memória: 2 etapas (recupera contexto do incidente anterior)

Redução de 60% nas chamadas = menor custo + menor latência
```

#### memory.md — Contrato de Memória

```yaml
memory:
  short_term:
    scope: "current_execution"
    includes: [current_step, decisions_made, tools_called, results]
    
  long_term:
    persistence: "file"  # ou "database"
    max_entries: 1000
    update_policy: "overwrite_if_exists"  # evita duplicação
    includes:
      - confirmed_system_facts
      - service_configurations
    excludes:
      - sensitive_data
      - unvalidated_assumptions
      
  episodic:
    persistence: "file"
    max_entries: 500
    includes:
      - execution_summaries
      - resolved_incidents
      - lessons_learned
      
  contextual:
    backend: "embeddings"
    similarity_threshold: 0.7
    max_retrieved_fragments: 5
    max_tokens_per_fragment: 200
```

#### Regras Críticas de Memória

```yaml
# rules.md — integrado com memória
memory_rules:
  - "NUNCA armazenar tokens de autenticação, senhas ou credenciais"
  - "APENAS armazenar fatos confirmados por ferramentas — nunca suposições"
  - "SE um fato muda no sistema → marcar como 'needs_revalidation'"
  - "ANTES de finalizar → persistir resumo da execução na memória episódica"
```

#### Agent Loop Evoluído

```
Antes (sem memória):
  Input → Percepção → Raciocínio → Ação → Avaliação → Output

Depois (com memória):
  Input → Percepção
              ↓
        Recuperar Contexto
        (long-term + episodic + contextual search)
              ↓
          Raciocínio (enriquecido com histórico)
              ↓
             Ação
              ↓
          Avaliação
              ↓
        Persistir Memória
        (fatos + resumo do episódio + lições)
              ↓
            Output
```

---

### Capítulo 2: Embeddings e Reflexão Evolutiva

#### Por que Embeddings Transformam a Memória

```python
# Busca por texto: falha em sinônimos
memory.search("latência alta no checkout")
# Resultado: [] (não encontra "tempo de resposta degradado")

# Busca por embedding: captura significado
embedding_search("latência alta no checkout", threshold=0.7)
# Resultado: [
#   {"text": "Tempo de resposta degradado causado por índice faltante", "score": 0.84},
#   {"text": "Alta latência no serviço de pagamentos", "score": 0.79}
# ]
```

#### Implementação do Embed Adapter

```python
class EmbedAdapter:
    def __init__(self, model_name: str = "text-embedding-3-small"):
        self.model = EmbeddingModel(model_name)
        self.vector_store = LocalVectorStore()  # arquivo JSON simples para volume baixo
    
    def index(self, text: str, metadata: dict) -> None:
        embedding = self.model.encode(text)
        self.vector_store.add({
            "embedding": embedding,
            "text": text,
            "metadata": metadata,
            "created_at": datetime.now().isoformat()
        })
    
    def search(self, query: str, threshold: float = 0.7, max_results: int = 5) -> list:
        query_embedding = self.model.encode(query)
        results = self.vector_store.similarity_search(query_embedding)
        return [r for r in results if r["score"] >= threshold][:max_results]
    
    def cosine_similarity(self, a: list, b: list) -> float:
        # Produto escalar / (norma_a × norma_b)
        dot_product = sum(x * y for x, y in zip(a, b))
        norm_a = sum(x ** 2 for x in a) ** 0.5
        norm_b = sum(x ** 2 for x in b) ** 0.5
        return dot_product / (norm_a * norm_b)
```

#### Reflexão Evolutiva — Extraindo Lições

```python
class ReflectionEngine:
    def extract_lessons(self, trace: Trace) -> list[Lesson]:
        # Só extrai lições se houve algo notável (erro ou insight)
        if trace.all_steps_success and trace.no_unexpected_behavior:
            return []  # execução normal não gera lição
        
        # LLM analisa o trace e extrai lições generalizáveis
        lessons = self.llm.analyze(
            trace=trace,
            prompt=self.reflection_prompt
        )
        
        # Filtra lições de qualidade
        return [l for l in lessons if self._is_high_quality(l)]
    
    def _is_high_quality(self, lesson: Lesson) -> bool:
        return (
            len(lesson.situation) > 20          # específica o suficiente
            and len(lesson.action) > 10          # acionável
            and lesson.generalizability in ["medium", "high"]  # reutilizável
            and not lesson.is_trivial            # não óbvia
        )
```

**Injeção de lições no planejamento:**
```python
# planner.py
def _build_prompt(self, perception: Perception) -> str:
    lessons = self.memory.get_relevant_lessons(perception.input, max=3)
    
    lessons_text = ""
    if lessons:
        lessons_text = "\n\nLESSONS LEARNED FROM PAST EXECUTIONS:\n"
        for lesson in lessons:
            lessons_text += f"- {lesson.text}\n"
    
    return f"""...{lessons_text}

Consider these lessons when planning your approach.
"""
```

**Diferença entre Memória e Aprendizado:**

| Aspecto | Memória | Reflexão Evolutiva |
|---|---|---|
| O que faz | Armazena o que aconteceu | Transforma o passado em diretriz |
| Resultado | "Lembro que o payment-api tem SLO de 200ms" | "Da próxima vez, verificar deploy antes de métricas" |
| Impacto | Evita reprocessamento | Muda a estratégia de decisão |

---

### Capítulo 3: Evals de Memória e Fechamento

#### Os Três Cenários de Impacto da Memória

```
Cenário 1 — Memória AJUDA:
  Sem memória: 5 etapas (descobre tudo do zero)
  Com memória: 3 etapas (usa contexto anterior para focar investigação)
  Decision Improvement: +40%

Cenário 2 — Memória é IRRELEVANTE:
  Sem memória: 4 etapas
  Com memória: 4 etapas (recupera fragmentos irrelevantes, descarta)
  Decision Improvement: 0% (mas custo de tokens aumenta!)

Cenário 3 — Memória PREJUDICA:
  Memória: "timeout padrão do payment-api é 30 segundos"
  Realidade atual: timeout foi reduzido para 10 segundos há 2 semanas
  Resultado: diagnóstico incorreto baseado em dado desatualizado
  Decision Improvement: -25%
```

> 🔑 **Memória não é garantia de melhoria.** Ela pode otimizar, ser neutra ou causar falhas. Sempre medir com evals.

#### As 6 Métricas de Memória

**Retrieval Precision:**
```python
# De todos os fragmentos recuperados, quantos eram realmente relevantes?
precision = relevant_retrieved / total_retrieved
# Se baixa → aumentar threshold de similaridade (0.7 → 0.8)
```

**Retrieval Recall:**
```python
# De todos os fragmentos relevantes existentes, quantos foram recuperados?
recall = relevant_retrieved / total_relevant_in_store
# Se baixo → diminuir threshold ou melhorar qualidade dos fragmentos indexados
```

**Memory Utilization:**
```python
# Do contexto recuperado, quanto o Planner efetivamente usou?
utilization = facts_used_in_decision / facts_provided_in_context
# Se baixo → problema no Planner (não está considerando o contexto)
```

**Hallucination from Memory:**
```python
# O agente inventou ou distorceu dados em relação à memória?
hallucination_rate = hallucinated_facts / total_memory_references
# Se alto → dados desatualizados → implementar TTL/expiração
```

**Decision Improvement (MÉTRICA CENTRAL):**
```python
# Comparação direta: com vs sem memória
improvement = (steps_with_memory - steps_without_memory) / steps_without_memory
# Negativo = memória PREJUDICA → analisar as outras métricas para causa raiz
```

**Lesson Quality:**
```python
# As lições extraídas são generalizáveis e acionáveis?
quality_score = sum(lesson.quality_score for lesson in lessons) / len(lessons)
# Se baixo → ajustar políticas de extração e critérios de qualidade
```

#### Ferramentas de Mercado para Evals de Memória

| Ferramenta | Especialidade |
|---|---|
| **Ragas** | Métricas de RAG: context_precision, context_recall, faithfulness, answer_relevancy |
| **Promptfoo** | Comparação de modelos/prompts/arquiteturas em escala |
| **LangSmith** | Observabilidade de traces LangChain + evals integrados |

#### Visão Geral do Sistema Completo

```
┌─────────────────────────────────────────────────────────────────┐
│                    AGENTE AUTÔNOMO COMPLETO                     │
├─────────────────────────────────────────────────────────────────┤
│  Contratos (agent.md, loop.md, planner.md, ...)                │
│  └── definem comportamento, limites, regras de segurança        │
│                                                                 │
│  Runtime (Python)                                               │
│  └── executa Agent Loop: Percepção → Raciocínio → Ação →       │
│       Avaliação → Persistência                                  │
│                                                                 │
│  Adapters (REST, DB, MCP)                                       │
│  └── conectam o agente a dados reais sem alterar contratos      │
│                                                                 │
│  Memória (curta, longa, episódica, contextual)                  │
│  └── acumula experiência e enriquece decisões futuras           │
│                                                                 │
│  Reflexão Evolutiva                                             │
│  └── extrai lições e melhora continuamente                      │
│                                                                 │
│  Evals (benchmark, tool selection, memória)                     │
│  └── mede comportamento com rigor científico                    │
│                                                                 │
│  Observabilidade (hooks, dashboard, trace, análise auto)        │
│  └── torna cada decisão rastreável e auditável                  │
└─────────────────────────────────────────────────────────────────┘
```

**Analogia final para devs Java:**
```
Modelo LLM       = Motor de decisão (não o agente em si)
Regras/Rules     = @PreAuthorize + Bean Validation + Circuit Breaker
Planner          = @Service que chama LLM API
Habilidades      = Interfaces de Repository/Service
Memória          = @Repository com múltiplas implementações
Reflexão         = Pipeline de analytics pós-execução
Trace            = OpenTelemetry Distributed Tracing para decisões
```

> ✅ **Conclusão:** A evolução não está no modelo. Está na **arquitetura**. O agente deixa de ser uma ferramenta e passa a se comportar como um sistema inteligente, capaz de tomar decisões cada vez melhores com o tempo — governado, observável, evolutivo e mensurável.
