# Decision Framework Tool — Java

Porte para Java 17 (Maven puro, **sem Spring Boot**) do protótipo original:

- `../decision-framework-tool.js` (versão oficial da disciplina)
- `../decision_framework_tool.py` (versão de referência espelhada)

Ambos os arquivos originais foram lidos por completo antes deste porte; a lógica é
idêntica nos dois (mesma árvore de três perguntas, mesmos quatro rótulos de
classificação, mesmos dois exemplos de demo). Não há divergência entre a versão JS e a
Python neste caso.

Contexto de domínio: `../decision-framework-checklist.md` (o checklist que este código
implementa) e `../ai-first-architecture-canvas.md` / `../reference-architecture-canvas.md`
(canvas do módulo, não portados — são artefatos de documentação/prompt, não código).

## O que é esta ferramenta

Lógica pura, 100% determinística: três perguntas booleanas (`p1`, `p2`, `p3`) viram uma
árvore de decisão simples que classifica uma tarefa em uma de quatro categorias. Não há
IA, não há modelo, não há chamada de rede — é o contraste pedagógico do módulo: quando a
decisão é auditável e finita, ela vira código com confiança total.

## Por que Maven puro, sem Spring Boot (e sem framework nenhum)

Este é um script/CLI standalone sem servidor HTTP — convenção já usada nos demais portes
de scripts do Módulo 01 (`ollama-local-llm-chat-java`, etc.). Especificamente para este
protótipo, Spring Boot foi avaliado e descartado: a ferramenta não tem nenhuma
integração externa, nenhum ponto de configuração/injeção de dependência que justifique um
container IoC, nem múltiplos componentes colaborando em runtime — é uma função pura
`(boolean, boolean, boolean) -> Classificacao` mais uma função de decomposição de lista.
Adicionar Spring Boot aqui adicionaria peso e tempo de boot sem nenhum ganho de clareza
ou testabilidade; `record`s imutáveis e um `enum` já resolvem tudo isso de forma mais
idiomática e mais simples de auditar.

## O que foi mantido 1:1

- A árvore de decisão (`classificarTarefa`): mesma ordem de curto-circuito (`p1` decide
  sozinho, senão `p2`, senão `p3`, senão o rótulo enumerável) e os mesmos quatro textos
  de classificação, caractere por caractere (incluindo acentuação em português).
- `decomporTarefaHibrida`: mesmo comportamento — aplica a árvore item a item sobre uma
  lista de subtarefas e devolve uma lista nova (não modifica a entrada), com o campo
  `classificacao` a mais.
- Os mesmos cenários de teste automatizado: as 4 combinações do Bloco 1 do checklist
  (incluindo os casos de curto-circuito, testando `p2`/`p3` "não importam" quando `p1` é
  verdadeiro, e `p3` "não importa" quando `p2` é verdadeiro) e a decomposição das duas
  subtarefas de referência do TrialForge (Bloco 2), com a mesma nota honesta do original
  sobre por que as outras duas linhas da tabela do checklist ficam fora do teste
  automatizado (misturam regra e gate condicional de um jeito que não mapeia limpo para
  uma única tripla `p1`/`p2`/`p3`).
- A mesma demo impressa no terminal: os 4 casos do Bloco 1 e a decomposição da tarefa
  híbrida do Bloco 2, com os mesmos rótulos e a mesma nota final.

## O que foi adaptado (e por quê)

- **Testes JUnit 5 em vez de `assert`/`unittest` embutido no mesmo comando**: no
  original, `node decision-framework-tool.js` (ou `python3 decision_framework_tool.py`)
  roda os testes automatizados e, se todos passarem, a demo, no mesmo comando/processo.
  Em Java, a convenção do projeto (mesma dos demais portes do repositório) separa testes
  automatizados (`mvn test`, JUnit 5) do binário de demo (`mvn exec:java`, classe
  `Main`). O comportamento observável de cada parte (testes e demo) é idêntico ao
  original; só a orquestração "testes antes da demo, no mesmo comando" não foi
  replicada — não há equivalente direto sem reinventar um test runner dentro de `Main`.
- **`Classificacao` como `enum`, não como objeto de constantes `String`**: o `enum`
  `toString()` devolve exatamente o mesmo texto que as versões JS/Python retornam como
  string, então o comportamento observável (o texto impresso/comparado) é idêntico. A
  troca ganha checagem de tipos em tempo de compilação, que não existe no JS/Python.
- **`Subtarefa`/`SubtarefaClassificada` como `record`s, não como `dict`/objeto
  literal**: a versão Python documenta explicitamente que `decompor_tarefa_hibrida` não
  modifica os dicts de entrada, devolvendo dicts novos. Em Java, um `record` é imutável
  por construção — a garantia de "não modifica a entrada" passa a ser do compilador, não
  só de disciplina de código, o que é mais idiomático em Java 17+.
- **Saída forçada para UTF-8 em `Main`**: adicionado `System.setOut(new PrintStream(...,
  StandardCharsets.UTF_8))` no início de `main()`. Isso não muda nenhum texto — só
  garante que os bytes emitidos sejam sempre UTF-8, independente do charset padrão da
  JVM/console (no Windows, o console frequentemente usa cp1252/cp850 por padrão em Java
  17, o que corrompia a acentuação em português na tela sem essa correção). Para ver os
  acentos corretamente no console do Windows, rode `chcp 65001` antes, ou redirecione a
  saída para um arquivo.

## Estrutura

```
decision-framework-tool-java/
├── pom.xml
├── src/main/java/com/decisionframework/
│   ├── Classificacao.java          # enum com as 4 classificações do checklist
│   ├── Subtarefa.java              # record: subtarefa ainda não classificada
│   ├── SubtarefaClassificada.java  # record: subtarefa + classificação
│   ├── DecisionFramework.java      # classificarTarefa() + decomporTarefaHibrida()
│   └── Main.java                   # demo (Bloco 1 e Bloco 2), stdout em UTF-8
└── src/test/java/com/decisionframework/
    └── DecisionFrameworkTest.java  # JUnit 5 + AssertJ, mesmos cenários do original
```

## Como rodar

Pré-requisitos: Java 17+ e Maven.

```bash
# Rodar os testes automatizados
mvn test

# Rodar a demo no terminal
mvn compile exec:java

# Ou empacotar e rodar o jar
mvn package
java -jar target/decision-framework-tool-1.0.0.jar
```

Nenhuma variável de ambiente, credencial ou chamada de rede é necessária — por isso não
há `.env.example` neste projeto.
