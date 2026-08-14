# Nexus Tools API em Go

Porte em Go das **ferramentas** (tools/*.py, decoradas com `@tool` do CrewAI) do projeto Nexus AIOps do módulo 06 do curso — expostas como endpoints HTTP simples. Ver também [`nexus-tools-java`](../nexus-tools-java), o porte Java irmão deste projeto.

**Escopo desta porta**: apenas a lógica de negócio das 18 ferramentas foi portada. A orquestração multi-agente hierárquica do CrewAI (`core/agents.py`, `labs/*.py`, o "Nexus Manager" que delega tarefas a agentes especialistas via LLM) **não** foi replicada — não existe framework equivalente pronto em Go, e reimplementá-la do zero seria um projeto de arquitetura à parte, fora do escopo de um porte mecânico.

## Endpoints

| Rota | Arquivo original | Descrição |
|---|---|---|
| `POST /tools/generate-k8s-manifest` | `k8s_ops.py` | Gera manifestos K8s Deployment+Service e salva em disco |
| `POST /tools/apply-k8s-manifest` | `k8s_ops.py` | Aplica manifesto via `kubectl` (ou simula se indisponível) |
| `POST /tools/analyze-canary-metrics` | `k8s_ops.py` | Decide proceed/rollback de um canary rollout |
| `POST /tools/inspect-pod-failure` | `k8s_diag.py` | Diagnostica CrashLoopBackOff/OOMKilled |
| `POST /tools/suggest-fix` | `k8s_diag.py` | Sugere remediação por tipo de problema |
| `POST /tools/run-checkov-scan` | `security_scan.py` | Roda o scanner Checkov via CLI (ou reporta ausência) |
| `POST /tools/validate-opa-policies` | `security_scan.py` | Simula o motor de políticas OPA (região, tamanho de instância, ingress público) |
| `POST /tools/query-prometheus-metrics` | `obs_tools.py` | Simula consulta PromQL |
| `POST /tools/query-jaeger-traces` | `obs_tools.py` | Simula consulta de tracing distribuído |
| `POST /tools/nl-to-promql` | `aiops_tools.py` | Traduz linguagem natural para PromQL/LogQL |
| `POST /tools/predictive-disk-alert` | `aiops_tools.py` | Simula alerta preditivo de saturação de disco |
| `POST /tools/generate-grafana-dashboard` | `aiops_tools.py` | Gera e salva um dashboard Grafana em JSON |
| `POST /tools/execute-terraform` | `chatops_tools.py` | Aplica Terraform com aprovação humana obrigatória para ações destrutivas |
| `POST /tools/triage-security-vulnerabilities` | `governance_tools.py` | Triagem de vulnerabilidades (Trivy/Snyk) |
| `POST /tools/optimize-cicd-pipeline` | `governance_tools.py` | Sugestões de otimização de pipeline CI/CD |
| `POST /tools/analyze-finops-costs` | `governance_tools.py` | Identifica recursos zumbis e sugere rightsizing |
| `POST /tools/check-compliance-rules` | `policy_rag.py` | Consulta políticas corporativas de nomenclatura/tagging |
| `POST /tools/write-file` | `file_writer.py` | Salva conteúdo em arquivo, removendo cercas markdown |
| `POST /tools/consult-runbook` | `labs/modulo10_remediation.py` | Lê o runbook oficial de um serviço (`data/runbook_<service>.md`) |

Todas as respostas seguem o envelope `{"result": "<texto>"}`.

## O que foi mantido 1:1

- Mesmo texto de resposta (incluindo emojis) para todos os ramos determinísticos/simulados de cada ferramenta.
- Mesmas regras de negócio: threshold de erro do canary, regras de governança OPA, aprovação humana obrigatória (`GESTOR-APROVA`) para comandos Terraform destrutivos, mapa de remediações conhecidas.
- Mesmo comportamento de integração externa: `apply-k8s-manifest` tenta `kubectl apply` de verdade e cai em modo de simulação se o binário não existir; `run-checkov-scan` tenta `checkov` da mesma forma.
- Mesma leitura de runbooks em `data/runbook_<service>.md`.

## O que foi adaptado

| Aspecto | Python/CrewAI | Go |
|---|---|---|
| Invocação da tool | Function calling do LLM dentro de um `Crew`/`Agent` do CrewAI | Endpoint HTTP `POST /tools/<nome>`, chamável por qualquer orquestrador (inclusive um agent loop já portado neste repositório, como [`agent-loop-framework-go`](../../modulo04-agentes-autonomos/agent-loop-framework-go)) |
| Execução de subprocessos | `subprocess.run(["kubectl", ...])` / `subprocess.run(["checkov", ...])` | `os/exec` |
| Geração/serialização de JSON (dashboard Grafana) | `json.dumps(dict, indent=2)` | `encoding/json` com `MarshalIndent` |

## Como executar

```bash
cp .env.example .env
go run .
```

```bash
curl -X POST http://localhost:8080/tools/check-compliance-rules -d '{"query":"nomenclatura de buckets"}'
curl -X POST http://localhost:8080/tools/consult-runbook -d '{"serviceName":"db"}'
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem os principais ramos de decisão de cada ferramenta (K8s ops/diag, security scan, observabilidade, AIOps, ChatOps, runbook), incluindo geração/leitura real de arquivos em disco, além de um smoke test HTTP.
