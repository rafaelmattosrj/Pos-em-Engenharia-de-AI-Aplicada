# Agent Components Demo — Go

Porte Go de [`agent-components-demo.js`](../agent-components-demo.js) (fonte primária,
usada na gravação) e [`agent_components_demo.py`](../agent_components_demo.py) (referência
idêntica em Python), do Módulo 2.1 — UNIPDS: Arquitetura de Sistemas com IA.

Cinco mini-demonstrações, uma por peça da anatomia de um agente único (Memória,
Planejamento, Ferramentas, Ação, Approval Gate) — cada uma isolada e rodável sozinha, sem
o loop ReAct inteiro (isso é [`react-agent-prototype-go`](../react-agent-prototype-go)) e
sem o schema formal de ferramenta. Contexto: TrialForge, Agente ICF (gera a seção de
assentimento do Termo de Consentimento a partir do protocolo do estudo).

## O que foi mantido 1:1

- As 5 seções e a ordem de execução: Memória → Planejamento → Ferramentas → Ação/Gate.
- `BuscarClausulaAssentimento`: mesma regra determinística (menor de 18 anos na faixa
  etária → cláusula ANVISA; só adultos → aviso).
- `ExecutarOuGatear`: mesmo comportamento — ação com `RequerAprovacao=true` nunca chega a
  rodar o `Executar()`.
- `ChainOfThought` / `ChainOfThoughtMaisReflexao`: mesma lógica de 1 vs. 2 chamadas ao
  modelo, e a mesma métrica de razão de tempo entre elas.
- Os mesmos 3 testes automatizados determinísticos do original (memória, ferramenta,
  gate), cobrindo os mesmos cenários.
- Mesmo texto de saída no console, incluindo os comentários pedagógicos (`->`).

## O que foi adaptado (e por quê)

- **Memória de longo prazo vira `struct` instanciável, não um mapa em escopo de pacote.**
  No original, `bancoDeUsuarios` é um `Map`/`dict` de módulo, compartilhado por todas as
  chamadas do processo — os testes chamam `.clear()` antes de rodar. Em Go isso viraria
  uma variável de pacote mutável (evitável e não idiomático para testes paralelos).
  `agente.MemoriaLongoPrazo` (criada com `NovaMemoriaLongoPrazo()`) é uma struct com seu
  próprio mapa por instância — mesmo comportamento observável (duas chamadas com o mesmo
  `usuarioID` acumulam estado), isolamento entre testes por instância nova em vez de
  `clear()`.
- **Chamada ao Ollama via `net/http` + `encoding/json`, direto na API nativa
  (`POST /api/chat`), sem SDK.** Mesmo padrão já usado em
  [`ollama-local-llm-chat-go`](../../../modulo01-fundamentos-de-ia-e-llms-para-programadores/ollama-local-llm-chat-go)
  (módulo 1), adaptado para o endpoint nativo de chat do Ollama em vez do endpoint
  OpenAI-compatible, já que o original chama `ollama.chat(...)` diretamente.
  `agente.ChatClient` é uma interface implementada por um adaptador sobre
  `internal/ollama.Client`, só para manter a chamada de rede isolada — a seção de
  Planejamento não tem teste automatizado, igual ao original (só observação ao vivo).
- **Objetos literais viram `struct`s.** `{texto, fonte, aviso}`, `{status, resultado,
  mensagem}`, `{role, content}` viram `agente.ResultadoClausula`, `agente.ResultadoAcao`,
  `agente.Mensagem` — tipados, idiomático em Go, mesmo formato observável.
- **`acaoProposta.executar` vira `func() string`** dentro da struct `AcaoProposta`, no
  lugar da função/closure solta do objeto literal original — mesmo efeito (só roda se não
  houver gate).

## Como rodar

Pré-requisito: Go 1.22+. A seção de Planejamento (única que chama o modelo de verdade)
precisa do [Ollama](https://ollama.com) rodando localmente com `ollama pull gemma4:e2b` —
as demais seções (Memória, Ferramentas, Ação/Gate) não dependem de rede e sempre rodam.

```bash
# Rodar a demo completa
go run .

# Rodar os testes automatizados (Memória, Ferramentas, Ação/Gate — sem rede)
go test ./...

# Verificação estática
go vet ./...
```

Copie `.env.example` para `.env` se quiser customizar `OLLAMA_BASE_URL` (o valor padrão já
é `http://localhost:11434`, sem precisar de arquivo `.env`).

## Estrutura

```
internal/agente/
  memoria.go          — Seção 1: memória curto/longo prazo
  planejamento.go     — Seção 2: chain-of-thought vs. reflexão (interface ChatClient)
  ferramentas.go       — Seção 3: BuscarClausulaAssentimento (determinístico)
  acaogate.go           — Seções 4/5: ExecutarOuGatear (determinístico)
  *_test.go             — testes dos três blocos determinísticos
internal/ollama/
  client.go             — cliente HTTP nativo do Ollama (POST /api/chat)
main.go                 — orquestra testes + demonstrações, na mesma ordem do original
dotenv.go                — leitura simples de .env (mesmo padrão de ollama-local-llm-chat-go)
```
