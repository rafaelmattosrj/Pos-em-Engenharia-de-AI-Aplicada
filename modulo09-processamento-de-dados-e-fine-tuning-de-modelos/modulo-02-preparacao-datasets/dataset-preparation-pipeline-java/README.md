# Dataset Preparation Pipeline (Java)

Porte Java das 5 ferramentas de `modulo-02-preparacao-datasets/` (originais em
JS/Python, ex.: `data-relevance-scoring-tool.js`). Agrupadas num unico
projeto Maven, organizado por pacote de responsabilidade, porque as 5
ferramentas formam um pipeline sequencial (relevancia -> extracao ->
limpeza/balanceamento -> PII) do mesmo dataset da Amplitude Seguros — 5
mini-projetos separados duplicariam o modelo de dados (`Exemplo`) sem
ganho real de isolamento.

## Origem de cada arquivo

| Pacote/classe | Original |
|---|---|
| `com.datasetprep.DataRelevance` | `data-relevance-scoring-tool.js` |
| `com.datasetprep.cleaning.*` | `dataset-cleaning-balancing-tool.js` (MinHash+LSH, temperatura, entropia) |
| `com.datasetprep.pii.PiiScrubbingGate` | `pii-scrubbing-gate-tool.js` |
| `com.datasetprep.extraction.OcrExtraction` | `extraction-to-jsonl-tool.js` |
| `com.datasetprep.extraction.LlmMultimodalExtraction` | `extracao-llm-multimodal-tool.js` |

## Mantido 1:1

- Todas as constantes de negocio (criterios de relevancia, candidatos reais
  da Amplitude Seguros, parametros de MinHash k=32/LSH b=8,r=4, alpha=0.3 de
  temperatura, regex de parsing tolerante a rotulo, regex de CPF/nome, e o
  algoritmo Modulo 11 de validacao de CPF).
- Os mesmos cenarios de teste do `.js` original (ver `src/test/java`).

## Adaptado (sem equivalente direto)

- **Nomes com acento**: "Clínica Vitalis" virou "Clinica Vitalis" nos ids
  internos do dataset simulado, para evitar problemas de encoding entre
  fontes Java/arquivo — nao muda nenhum resultado dos testes, so o literal
  do id.
- **Chamada ao Tesseract** (`OcrExtraction.ocrTexto`/`ocrConfiancaMedia`):
  portada via `ProcessBuilder` invocando o binario `tesseract` no PATH
  (equivalente a `execFileSync` no original). Requer Tesseract instalado
  com o pacote de idioma `por` — **nao roda em CI/teste automatizado sem
  essa dependencia externa**; por isso as funcoes de parsing/validacao de
  schema (que sao puras) tem testes proprios, e a chamada ao binario nao.
- **Chamada ao Vertex AI/Gemini** (`LlmMultimodalExtraction.extrairViaLlm`):
  portada com `java.net.http.HttpClient` (nativo, sem lib externa) +
  `ProcessBuilder` para obter o token via `gcloud auth print-access-token`.
  Requer `gcloud auth login` feito nesta maquina e acesso ao projeto GCP —
  **nao roda em CI/teste automatizado**. O parsing da resposta JSON usa um
  parser minimo baseado em regex (sem dependencia de lib JSON externa, para
  manter o projeto Maven puro/sem Spring Boot conforme a convencao de
  CLI standalone da skill de porte).
- Sem Spring Boot, sem LangChain4j: nenhuma das 5 ferramentas expoe servidor
  HTTP nem usa abstracao de LLM alem de uma chamada HTTP direta.

## Rodar

```bash
mvn compile test   # roda os testes offline (relevancia, dedup/balanceamento/diversidade, PII, parsing puro)
mvn exec:java      # roda o demo offline (Main.java)
```

Os demos de extracao real (OCR via Tesseract, LLM multimodal via Vertex AI)
nao tem `main` própria fiada no `exec-maven-plugin` — chame
`OcrExtraction`/`LlmMultimodalExtraction` a partir do seu proprio codigo,
com o binario/credencial correspondente disponivel.
