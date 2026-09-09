// Package dolly e' o extra "dataset real alternativo" (databricks-dolly-15k,
// CC-BY-SA-3.0) -- porte de dolly-dataset-real-starter.js e
// dolly-vertex-pipeline.js / DollyDatasetStarter.java e DollyVertexPipeline.java.
// Conceitualmente equivalente ao Modulo 2.2 (preparacao de dataset), so' que
// contra dado real, nao sintetico. Requer baixar o arquivo (13MB, ~15 mil
// linhas JSONL) antes de rodar de verdade -- ver README deste projeto.
package dolly

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"fine-tuning-via-api/datasetscaling"
	"fine-tuning-via-api/gemini"
	"fine-tuning-via-api/hyperparam"
	"fine-tuning-via-api/minhash"
)

// Caso e' o identificador de caso usado pra todo exemplo mapeado do Dolly.
const Caso = "dolly-instruction-tuning"

// NShingle e' o tamanho de shingle usado na deduplicacao deste extra (5
// palavras) -- igual ao usado no pipeline principal de DatasetScaling.
const NShingle = 5

// CategoriasCompativeis sao as categorias do Dolly-15k equivalentes as
// tarefas de extracao estruturada da Amplitude Seguros.
var CategoriasCompativeis = []string{"information_extraction", "closed_qa", "summarization"}

// RegistroDolly e' uma linha do databricks-dolly-15k.jsonl.
type RegistroDolly struct {
	Instruction string `json:"instruction"`
	Context     string `json:"context"`
	Response    string `json:"response"`
	Category    string `json:"category"`
}

// CarregarDolly lê o arquivo JSONL linha a linha -- equivalente a
// DollyDatasetStarter.carregarDolly.
func CarregarDolly(caminhoJsonl string) ([]RegistroDolly, error) {
	f, err := os.Open(caminhoJsonl)
	if err != nil {
		return nil, fmt.Errorf("dolly: falha ao abrir %s: %w", caminhoJsonl, err)
	}
	defer f.Close()

	var registros []RegistroDolly
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())
		if linha == "" {
			continue
		}
		var r RegistroDolly
		if err := json.Unmarshal([]byte(linha), &r); err != nil {
			return nil, fmt.Errorf("dolly: linha invalida: %w", err)
		}
		registros = append(registros, r)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("dolly: falha ao ler %s: %w", caminhoJsonl, err)
	}
	return registros, nil
}

func categoriaCompativel(categoria string) bool {
	for _, c := range CategoriasCompativeis {
		if c == categoria {
			return true
		}
	}
	return false
}

// FiltrarCompativeis mantém só registros de categoria compatível, com
// contexto e resposta não vazios -- equivalente a
// DollyDatasetStarter.filtrarCompativeis.
func FiltrarCompativeis(registros []RegistroDolly) []RegistroDolly {
	var compativeis []RegistroDolly
	for _, r := range registros {
		if categoriaCompativel(r.Category) && strings.TrimSpace(r.Context) != "" && strings.TrimSpace(r.Response) != "" {
			compativeis = append(compativeis, r)
		}
	}
	return compativeis
}

// ParaSchemaCanonico mapeia registros Dolly pro schema canônico compartilhado
// com datasetscaling.ExemploDataset -- equivalente a
// DollyDatasetStarter.paraSchemaCanonico. "fonte" aqui é a categoria Dolly
// (não um nome de oficina/clínica, como em datasetscaling).
func ParaSchemaCanonico(registros []RegistroDolly) []datasetscaling.ExemploDataset {
	mapeados := make([]datasetscaling.ExemploDataset, len(registros))
	for i, r := range registros {
		mapeados[i] = datasetscaling.ExemploDataset{
			Instrucao: r.Instruction,
			Entrada:   r.Context,
			Saida:     map[string]any{"texto": r.Response},
			Caso:      Caso,
			Fonte:     r.Category,
			ID:        fmt.Sprintf("dolly-%s-%d", r.Category, i),
		}
	}
	return mapeados
}

// textoParaDedup concatena instrução+entrada -- diferente do critério do
// pipeline principal (datasetscaling usa só "entrada"): aqui a instrução
// varia por exemplo (cada registro Dolly tem sua própria instrução), então
// concatenar discrimina; lá a instrução é fixa por caso, então concatenar só
// infla a similaridade. Ver nota em datasetscaling.ExemploDataset.
func textoParaDedup(e datasetscaling.ExemploDataset) string {
	return e.Instrucao + "\n" + e.Entrada
}

// ResultadoPreparo é o relatório do pipeline completo de preparação --
// equivalente a DollyDatasetStarter.ResultadoPreparo.
type ResultadoPreparo struct {
	Bruto          int
	Compativeis    int
	Mapeados       int
	ItensRemovidos int
	SemDuplicatas  int
	Contagem       map[string]int
	Alocacao       map[string]int
	Balanceado     int
}

// PrepararDatasetCompleto roda o pipeline completo (dedup no dataset inteiro
// + balanceamento por temperatura) -- equivalente a
// DollyDatasetStarter.prepararDatasetCompleto.
func PrepararDatasetCompleto(caminhoJsonl string, alvoTotal int) (ResultadoPreparo, error) {
	bruto, err := CarregarDolly(caminhoJsonl)
	if err != nil {
		return ResultadoPreparo{}, err
	}
	compativeis := FiltrarCompativeis(bruto)
	mapeados := ParaSchemaCanonico(compativeis)

	paraDedup := make([]minhash.Exemplo, len(mapeados))
	for i, e := range mapeados {
		paraDedup[i] = minhash.Exemplo{ID: e.ID, Caso: e.Caso, Fonte: e.Fonte, TextoParaDedup: textoParaDedup(e)}
	}
	dedup := minhash.EncontrarQuaseDuplicatasGenerico(paraDedup, NShingle)
	remover := make(map[int]bool, len(dedup.ParesDuplicata))
	for _, par := range dedup.ParesDuplicata {
		remover[par.J] = true
	}

	var semDuplicatas []datasetscaling.ExemploDataset
	for i, e := range mapeados {
		if !remover[i] {
			semDuplicatas = append(semDuplicatas, e)
		}
	}

	contagem := make(map[string]int)
	for _, e := range semDuplicatas {
		contagem[e.Fonte]++
	}
	alocacao := minhash.AlocarComCapacidade(contagem, minhash.AlphaTemperatura, alvoTotal)

	usados := make(map[string]int)
	balanceado := 0
	for _, e := range semDuplicatas {
		usadoAtual := usados[e.Fonte]
		if usadoAtual < alocacao[e.Fonte] {
			usados[e.Fonte] = usadoAtual + 1
			balanceado++
		}
	}

	return ResultadoPreparo{
		Bruto: len(bruto), Compativeis: len(compativeis), Mapeados: len(mapeados),
		ItensRemovidos: len(remover), SemDuplicatas: len(semDuplicatas),
		Contagem: contagem, Alocacao: alocacao, Balanceado: balanceado,
	}, nil
}

// ConverterParaGemini converte um exemplo Dolly já mapeado pro formato Gemini
// usando gemini.ConverterTexto -- a saída do Dolly é texto solto (chave
// "texto" no map Saida), não objeto estruturado como no caso Amplitude.
func ConverterParaGemini(e datasetscaling.ExemploDataset) (gemini.Conversao, error) {
	texto, _ := e.Saida["texto"].(string)
	return gemini.ConverterTexto(e.Instrucao, e.Entrada, texto)
}

// ConfigPipeline sao os parametros de um job de fine-tuning do extra Dolly --
// equivalente a DollyVertexPipeline.ConfigPipeline.
type ConfigPipeline struct {
	CaminhoLocal           string
	URIDataset             string
	BaseModel              string
	DisplayName            string
	EpochCount             int
	LearningRateMultiplier float64
}

// ResultadoPipeline e' o resultado de RodarPipeline.
type ResultadoPipeline struct {
	JobCriado map[string]any
	JobFinal  map[string]any
}

// UploadFn, CriarJobFn e AcompanharFn sao os mesmos pontos de injecao usados
// em FineTuningAutomation (Java: UploadFn/CriarJobFn/AcompanharFn) -- a
// implementacao real toca rede/gcloud, a de teste devolve estado fixo.
type UploadFn func(caminhoLocal, uriGcs string) error
type CriarJobFn func(config map[string]any, confirmar bool) (map[string]any, error)
type AcompanharFn func(nomeJob string, aoAtualizar func(map[string]any)) (map[string]any, error)

// RodarPipeline sobe o dataset Dolly já preparado, cria o job real e
// acompanha -- equivalente a DollyVertexPipeline.rodarPipeline. Reusa a
// mesma trava de confirmação/validação de hiperparâmetro dos demais pacotes
// (hyperparam.Validar bloqueia ANTES de qualquer chamada de rede).
func RodarPipeline(
	config ConfigPipeline, confirmar bool,
	uploadFn UploadFn, criarJobFn CriarJobFn, acompanharFn AcompanharFn,
	aoAtualizar func(job map[string]any),
) (ResultadoPipeline, error) {
	if err := hyperparam.Validar(hyperparam.Hiperparametros{
		EpochCount:             config.EpochCount,
		LearningRateMultiplier: config.LearningRateMultiplier,
	}); err != nil {
		return ResultadoPipeline{}, err
	}
	if err := uploadFn(config.CaminhoLocal, config.URIDataset); err != nil {
		return ResultadoPipeline{}, err
	}
	jobCriado, err := criarJobFn(map[string]any{}, confirmar)
	if err != nil {
		return ResultadoPipeline{}, err
	}
	jobFinal, err := acompanharFn(fmt.Sprint(jobCriado["name"]), aoAtualizar)
	if err != nil {
		return ResultadoPipeline{}, err
	}
	return ResultadoPipeline{JobCriado: jobCriado, JobFinal: jobFinal}, nil
}
