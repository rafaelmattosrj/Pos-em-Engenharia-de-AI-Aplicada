# Decision Framework Tool — Go

Porte para Go (módulo simples, `go.mod`, **sem framework HTTP**) do protótipo original:

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

## Por que sem framework HTTP (`net/http` puro nem chega a ser usado)

Este é um script/CLI standalone sem servidor HTTP — convenção já usada nos demais portes
de scripts do Módulo 01 (`ollama-local-llm-chat-go`, etc.). Frameworks de rota como
`chi`/`gin`/`echo` foram avaliados e descartados aqui: a ferramenta não expõe nenhuma
rota, não recebe requisição nenhuma — é uma função pura `(bool, bool, bool) ->
Classificacao` mais uma função de decomposição de slice, chamada diretamente por
`main()`. Não há superfície HTTP para um framework de rotas resolver.

## O que foi mantido 1:1

- A árvore de decisão (`ClassificarTarefa`): mesma ordem de curto-circuito (`p1` decide
  sozinho, senão `p2`, senão `p3`, senão o rótulo enumerável) e os mesmos quatro textos
  de classificação, caractere por caractere (incluindo acentuação em português).
- `DecomporTarefaHibrida`: mesmo comportamento — aplica a árvore item a item sobre um
  slice de subtarefas e devolve um slice novo (não modifica a entrada), com o campo
  `Classificacao` a mais.
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

- **Testes com `go test` em vez de rodarem no mesmo comando da demo**: no original,
  `node decision-framework-tool.js` (ou `python3 decision_framework_tool.py`) roda os
  testes automatizados e, se todos passarem, a demo, no mesmo comando/processo. Em Go, a
  convenção do projeto (mesma dos demais portes do repositório) separa testes
  automatizados (`go test ./...`) do binário de demo (`go run .`). O comportamento
  observável de cada parte (testes e demo) é idêntico ao original; só a orquestração
  "testes antes da demo, no mesmo comando" não foi replicada — não há equivalente direto
  sem reinventar um test runner dentro de `main()`.
- **`Classificacao` como tipo `string` nomeado com constantes, não como objeto de
  constantes livre**: mais idiomático em Go que usar `string` cru; o valor impresso ou
  comparado é exatamente o mesmo texto das versões JS/Python.
- **`Subtarefa`/`SubtarefaClassificada` como `struct`s, não como `map`/objeto literal**:
  campos nomeados e tipados, mais idiomático em Go que um `map[string]any` genérico
  (que seria o análogo mais literal do dict Python). `DecomporTarefaHibrida` constrói um
  slice novo a cada chamada, preservando o comportamento documentado no original de não
  modificar a entrada.
- **Pacote `decision` separado de `main`**: a lógica de negócio (árvore de decisão +
  decomposição) fica isolada em `decision/`, testável isoladamente, seguindo o mesmo
  padrão de organização por responsabilidade já usado em outros portes Go do módulo 01
  (ex.: `ollama-local-llm-chat-go/ollama/`).

## Estrutura

```
decision-framework-tool-go/
├── go.mod
├── main.go                          # demo (Bloco 1 e Bloco 2)
└── decision/
    ├── classifier.go                # ClassificarTarefa() + DecomporTarefaHibrida()
    └── classifier_test.go           # mesmos cenários de teste do original
```

## Como rodar

Pré-requisitos: Go 1.22+.

```bash
# Compilar
go build ./...

# Analise estática
go vet ./...

# Rodar os testes automatizados
go test ./...

# Rodar a demo no terminal
go run .
```

Nenhuma variável de ambiente, credencial ou chamada de rede é necessária — por isso não
há `.env.example` neste projeto.

Nota sobre acentuação no console do Windows: a saída é sempre UTF-8 (padrão em Go); se o
terminal não estiver configurado para UTF-8, rode `chcp 65001` antes, ou use um terminal
já configurado para UTF-8 (Windows Terminal, git-bash, etc.).
