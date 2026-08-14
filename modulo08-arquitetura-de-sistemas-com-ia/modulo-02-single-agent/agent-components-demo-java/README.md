# Agent Components Demo — Java

Porte Java de [`agent-components-demo.js`](../agent-components-demo.js) (fonte primária,
usada na gravação) e [`agent_components_demo.py`](../agent_components_demo.py) (referência
idêntica em Python), do Módulo 2.1 — UNIPDS: Arquitetura de Sistemas com IA.

Cinco mini-demonstrações, uma por peça da anatomia de um agente único (Memória,
Planejamento, Ferramentas, Ação, Approval Gate) — cada uma isolada e rodável sozinha, sem
o loop ReAct inteiro (isso é [`react-agent-prototype`](../react-agent-prototype-java)) e
sem o schema formal de ferramenta. Contexto: TrialForge, Agente ICF (gera a seção de
assentimento do Termo de Consentimento a partir do protocolo do estudo).

## O que foi mantido 1:1

- As 5 seções e a ordem de execução: Memória → Planejamento → Ferramentas → Ação/Gate.
- `buscarClausulaAssentimento` (`Ferramentas.buscarClausulaAssentimento`): mesma regra
  determinística (menor de 18 anos na faixa etária → cláusula ANVISA; só adultos → aviso).
- `executarOuGatear` (`AcaoGate.executarOuGatear`): mesmo comportamento — ação com
  `requerAprovacao=true` nunca chega a rodar o `executar()`.
- `chainOfThought` / `chainOfThoughtMaisReflexao` (`Planejamento`): mesma lógica de 1 vs. 2
  chamadas ao modelo, e a mesma métrica de razão de tempo entre elas.
- Os mesmos 3 testes automatizados determinísticos do original (`testarMemoria`,
  `testarFerramenta`, `testarGate`), cobrindo os mesmos cenários.
- Mesmo texto de saída no console, incluindo os comentários pedagógicos (`->`).

## O que foi adaptado (e por quê)

- **Memória de longo prazo vira objeto instanciável, não estado global estático.** No
  original, `bancoDeUsuarios` é um `Map`/`dict` em escopo de módulo, compartilhado por
  todas as chamadas do processo — os testes chamam `.clear()` antes de rodar para isolar
  cenários. Em Java isso viraria um campo `static` mutável, um anti-padrão evitável aqui:
  `Memoria.MemoriaLongoPrazo` é uma classe instanciável com seu próprio mapa por
  instância. O comportamento observável (duas chamadas com o mesmo `usuarioId` acumulam
  estado) é idêntico; só a forma de isolar entre testes muda (nova instância em vez de
  `clear()`).
- **Chamada ao Ollama via `java.net.http.HttpClient` + Jackson, direto na API nativa
  (`POST /api/chat`), sem SDK.** O original usa o pacote npm/pip `ollama`, que é um
  wrapper fino sobre essa mesma API REST. Em vez de puxar uma dependência de alto nível
  (LangChain4j) para uma única chamada de chat simples, replicamos a chamada HTTP
  diretamente — mesmo padrão já usado no projeto irmão
  [`manipulation-guardrail-prototype-java`](../../modulo-05-arquitetura-enterprise/manipulation-guardrail-prototype-java)
  deste módulo. `ChatClient` é uma interface (implementada por `OllamaChatClient`) só para
  manter a chamada de rede isolada e substituível, não porque os testes precisem de um
  dublê (a seção de Planejamento não é testada automaticamente, igual ao original).
- **Objetos literais viram `record`s.** `{texto, fonte, aviso}`, `{status, resultado,
  mensagem}`, `{role, content}` viram `Ferramentas.ResultadoClausula`,
  `AcaoGate.ResultadoAcao`, `Memoria.Mensagem` — tipados e imutáveis, idiomático em Java
  17, mesmo formato observável.
- **`acaoProposta.executar` vira `Supplier<String>`** dentro do record `AcaoProposta`, no
  lugar da função/lambda solta do objeto literal original — mesmo efeito (só roda se não
  houver gate).

## Como rodar

Pré-requisito: Java 17+ e Maven. A seção de Planejamento (única que chama o modelo de
verdade) precisa do [Ollama](https://ollama.com) rodando localmente com
`ollama pull gemma4:e2b` — as demais seções (Memória, Ferramentas, Ação/Gate) não
dependem de rede e sempre rodam.

```bash
# Compilar e rodar a demo completa
mvn compile exec:java

# Rodar os testes automatizados (Memória, Ferramentas, Ação/Gate — sem rede)
mvn test
```

Copie `.env.example` para `.env` se quiser customizar `OLLAMA_BASE_URL` (o valor padrão já
é `http://localhost:11434`, sem precisar de arquivo `.env`).

## Estrutura

```
src/main/java/com/trialforge/agentcomponents/
  Memoria.java          — Seção 1: memória curto/longo prazo
  Planejamento.java     — Seção 2: chain-of-thought vs. reflexão (chama o Ollama)
  Ferramentas.java      — Seção 3: buscarClausulaAssentimento (determinístico)
  AcaoGate.java          — Seções 4/5: executarOuGatear (determinístico)
  ChatClient.java        — abstração da chamada ao modelo
  OllamaChatClient.java  — implementação real via HTTP nativo do Ollama
  Main.java              — orquestra testes + demonstrações, na mesma ordem do original
src/test/java/com/trialforge/agentcomponents/
  MemoriaTest.java
  FerramentasTest.java
  AcaoGateTest.java
```
