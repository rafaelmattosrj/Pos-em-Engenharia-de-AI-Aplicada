# Dataset Preparation Pipeline (Go)

Porte Go das 5 ferramentas de `modulo-02-preparacao-datasets/` (originais em
JS/Python, ex.: `data-relevance-scoring-tool.js`). Agrupadas num unico
modulo Go, organizado por pacote de responsabilidade — irmao do porte Java
(`../dataset-preparation-pipeline-java/`), mesma decisao de agrupamento:
as 5 ferramentas formam um pipeline sequencial (relevancia -> extracao ->
limpeza/balanceamento -> PII) do mesmo dataset da Amplitude Seguros.

## Origem de cada pacote

| Pacote | Original |
|---|---|
| `datarelevance` | `data-relevance-scoring-tool.js` |
| `cleaning` | `dataset-cleaning-balancing-tool.js` (MinHash+LSH, temperatura, entropia) |
| `pii` | `pii-scrubbing-gate-tool.js` |
| `extraction` (ocr.go) | `extraction-to-jsonl-tool.js` |
| `extraction` (llmmultimodal.go) | `extracao-llm-multimodal-tool.js` |

## Mantido 1:1

- Todas as constantes de negocio, candidatos reais da Amplitude Seguros,
  parametros de MinHash k=32/LSH b=8,r=4, alpha=0.3 de temperatura, regex de
  parsing tolerante a rotulo, regex de CPF/nome, e o algoritmo Modulo 11 de
  validacao de CPF.
- Os mesmos cenarios de teste do `.js` original (ver arquivos `_test.go`).

## Adaptado (sem equivalente direto)

- **Nomes com acento**: "Clínica Vitalis" virou "Clinica Vitalis" nos ids
  internos do dataset simulado (mesma adaptacao do porte Java, por
  consistencia entre os dois).
- **Remocao de acento** (`extraction.Normalizar`): implementada com um
  `strings.Replacer` manual para o alfabeto portugues, em vez de trazer a
  dependencia externa `golang.org/x/text` so para normalizacao Unicode —
  mantem o modulo sem dependencias externas, coerente com o padrao "Go
  simples, sem framework" da convencao de porte para CLI standalone.
- **Chamada ao Tesseract** (`extraction.OcrTexto`/`OcrConfiancaMedia`):
  portada via `os/exec.Command` chamando o binario `tesseract` no PATH.
  Requer Tesseract instalado com o pacote de idioma `por` — **nao roda em
  CI/teste automatizado sem essa dependencia externa**; por isso as funcoes
  de parsing/validacao de schema (puras) tem teste proprio, e a chamada ao
  binario nao.
- **Chamada ao Vertex AI/Gemini** (`extraction.ExtrairViaLlm`): portada com
  `net/http` (stdlib) + `os/exec` para obter o token via
  `gcloud auth print-access-token`. Requer `gcloud auth login` feito nesta
  maquina e acesso ao projeto GCP — **nao roda em CI/teste automatizado**.
- Sem framework web (`chi`/`gin`/`echo`): nenhuma das 5 ferramentas expoe
  servidor HTTP.

## Rodar

```bash
go build ./...   # compila
go vet ./...     # lint estatico
go test ./...    # roda os testes offline (relevancia, dedup/balanceamento/diversidade, PII, parsing puro)
go run .          # roda o demo offline (main.go)
```

Os demos de extracao real (OCR via Tesseract, LLM multimodal via Vertex AI)
nao tem `main` propria — chame `extraction.OcrTexto`/`extraction.ExtrairViaLlm`
a partir do seu proprio codigo, com o binario/credencial correspondente
disponivel.
