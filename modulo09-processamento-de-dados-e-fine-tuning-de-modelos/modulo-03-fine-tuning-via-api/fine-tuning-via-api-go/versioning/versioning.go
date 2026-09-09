// Package versioning implementa o versionamento e documentacao de modelo
// fine-tunado (Modulo 3.5): hash de conteudo do dataset (SHA-256, mesmo
// principio de git/Docker), ficha de versionamento a partir do job real, e
// model card em Markdown.
package versioning

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

const TaxaRealPorUnidade = 0.00002909

type FaixaCustoGpu struct {
	ConsumerMin, ConsumerMax, H100Min, H100Max float64
}

var FaixaCustoGpuPadrao = FaixaCustoGpu{ConsumerMin: 0.40, ConsumerMax: 0.80, H100Min: 2.50, H100Max: 4.00}

func CalcularHashDataset(caminho string) (string, error) {
	conteudo, err := os.ReadFile(caminho)
	if err != nil {
		return "", err
	}
	soma := sha256.Sum256(conteudo)
	return hex.EncodeToString(soma[:]), nil
}

func DuracaoSegundos(criadoEmISO, concluidoEmISO string) (float64, error) {
	inicio, err := time.Parse(time.RFC3339Nano, criadoEmISO)
	if err != nil {
		return 0, err
	}
	fim, err := time.Parse(time.RFC3339Nano, concluidoEmISO)
	if err != nil {
		return 0, err
	}
	return fim.Sub(inicio).Seconds(), nil
}

func FormatarDuracaoComEspaco(duracaoSegundos float64) string {
	minutos := math.Floor(duracaoSegundos / 60)
	segundos := math.Round(duracaoSegundos - minutos*60)
	return fmt.Sprintf("%dmin %ds", int64(minutos), int64(segundos))
}

type CustoEstimado struct {
	DuracaoSegundos                                    float64
	DuracaoFormatada                                    string
	ConsumerMinUsd, ConsumerMaxUsd, H100MinUsd, H100MaxUsd float64
}

func CalcularCustoEstimado(duracaoSegundos float64, faixa FaixaCustoGpu) CustoEstimado {
	duracaoHoras := duracaoSegundos / 3600
	return CustoEstimado{
		DuracaoSegundos:  math.Round(duracaoSegundos*1000) / 1000,
		DuracaoFormatada: FormatarDuracaoComEspaco(duracaoSegundos),
		ConsumerMinUsd:   arredondar2(duracaoHoras * faixa.ConsumerMin),
		ConsumerMaxUsd:   arredondar2(duracaoHoras * faixa.ConsumerMax),
		H100MinUsd:       arredondar2(duracaoHoras * faixa.H100Min),
		H100MaxUsd:       arredondar2(duracaoHoras * faixa.H100Max),
	}
}

type CustoReal struct {
	Unidades   int64
	CustoReais float64
}

func CalcularCustoReal(tokensCobraveis, epochCount *int64) *CustoReal {
	if tokensCobraveis == nil || epochCount == nil {
		return nil
	}
	unidades := *tokensCobraveis * *epochCount
	return &CustoReal{Unidades: unidades, CustoReais: math.Round(float64(unidades)*TaxaRealPorUnidade*100) / 100}
}

func arredondar2(v float64) float64 { return math.Round(v*100) / 100 }

type FichaVersionamento struct {
	JobID, ModeloBase, NomeExibicao, DatasetUri, DatasetHashSha256 string
	EpochCount                                                     *int64
	LearningRateMultiplier                                         *float64
	AdapterSize                                                    string
	ExemplosDataset, TokensCobraveis                                *int64
	ModeloAjustado, Endpoint, CriadoEm, ConcluidoEm, Estado         string
	CustoEstimado                                                   *CustoEstimado
	CustoReal                                                       *CustoReal
}

func mapaAninhado(m map[string]any, chave string) map[string]any {
	if v, ok := m[chave].(map[string]any); ok {
		return v
	}
	return map[string]any{}
}

func stringOu(m map[string]any, chave string) string {
	if v, ok := m[chave].(string); ok {
		return v
	}
	return ""
}

func int64Ptr(m map[string]any, chave string) *int64 {
	v, ok := m[chave]
	if !ok || v == nil {
		return nil
	}
	if f, ok := v.(float64); ok {
		n := int64(f)
		return &n
	}
	return nil
}

func float64Ptr(m map[string]any, chave string) *float64 {
	v, ok := m[chave]
	if !ok || v == nil {
		return nil
	}
	if f, ok := v.(float64); ok {
		return &f
	}
	return nil
}

func GerarFichaVersionamento(job map[string]any, hashDataset string) (FichaVersionamento, error) {
	jobID := stringOu(job, "name")
	if jobID == "" {
		return FichaVersionamento{}, errors.New(`job inválido: precisa ter ao menos "name"`)
	}

	tuningSpec := mapaAninhado(job, "supervisedTuningSpec")
	hiper := mapaAninhado(tuningSpec, "hyperParameters")
	stats := mapaAninhado(mapaAninhado(job, "tuningDataStats"), "supervisedTuningDataStats")
	tunedModel := mapaAninhado(job, "tunedModel")

	var custoEstimado *CustoEstimado
	createTime := stringOu(job, "createTime")
	endTime := stringOu(job, "endTime")
	if createTime != "" && endTime != "" {
		dur, err := DuracaoSegundos(createTime, endTime)
		if err != nil {
			return FichaVersionamento{}, err
		}
		ce := CalcularCustoEstimado(dur, FaixaCustoGpuPadrao)
		custoEstimado = &ce
	}

	tokensCobraveis := int64Ptr(stats, "totalBillableTokenCount")
	epochCount := int64Ptr(hiper, "epochCount")
	custoReal := CalcularCustoReal(tokensCobraveis, epochCount)

	return FichaVersionamento{
		JobID: jobID, ModeloBase: stringOu(job, "baseModel"), NomeExibicao: stringOu(job, "tunedModelDisplayName"),
		DatasetUri: stringOu(tuningSpec, "trainingDatasetUri"), DatasetHashSha256: hashDataset,
		EpochCount: epochCount, LearningRateMultiplier: float64Ptr(hiper, "learningRateMultiplier"),
		AdapterSize:     stringOu(hiper, "adapterSize"),
		ExemplosDataset: int64Ptr(stats, "tuningDatasetExampleCount"), TokensCobraveis: tokensCobraveis,
		ModeloAjustado: stringOu(tunedModel, "model"), Endpoint: stringOu(tunedModel, "endpoint"),
		CriadoEm: createTime, ConcluidoEm: endTime, Estado: stringOu(job, "state"),
		CustoEstimado: custoEstimado, CustoReal: custoReal,
	}, nil
}

func ValidarFichaCompleta(f FichaVersionamento) error {
	var faltando []string
	if f.JobID == "" {
		faltando = append(faltando, "jobId")
	}
	if f.ModeloBase == "" {
		faltando = append(faltando, "modeloBase")
	}
	if f.DatasetUri == "" {
		faltando = append(faltando, "datasetUri")
	}
	if f.DatasetHashSha256 == "" {
		faltando = append(faltando, "datasetHashSha256")
	}
	if f.ModeloAjustado == "" {
		faltando = append(faltando, "modeloAjustado")
	}
	if f.Endpoint == "" {
		faltando = append(faltando, "endpoint")
	}
	if len(faltando) > 0 {
		return fmt.Errorf("Ficha de versionamento incompleta, faltam: %s", strings.Join(faltando, ", "))
	}
	return nil
}

type JobDpoReal struct {
	JobID, Estado, DuracaoFormatada string
	Exemplos                        int
}

var JobDpoRealAtual = JobDpoReal{
	JobID: "tuningJobs/3733013646142341120", Estado: "JOB_STATE_SUCCEEDED", DuracaoFormatada: "17min 32s", Exemplos: 40,
}

func GerarSecaoDpo() string {
	j := JobDpoRealAtual
	return fmt.Sprintf(`
## Continuação: preference tuning (DPO)

O Módulo 3.5 vai além do fine-tuning supervisionado acima e testa preference tuning (DPO) sobre o mesmo modelo base: 40 dos 200 exemplos deste job foram convertidos em pares de preferência (`+"`chosen`/`rejected`"+`) - a extração correta de sempre (`+"`chosen`"+`) contra uma resposta real gerada por um prompt deliberadamente mais fraco (`+"`rejected`"+`, sem exigir JSON estrito). O dataset de preferência resultante está em `+"`preference-dataset-amplitude.jsonl`"+`, nesta mesma pasta.

- Job: `+"`%s`"+`
- Estado: %s
- Duração real: %s (mais rápido que o SFT acima, dataset 5x menor: %d exemplos contra 200)
- Exemplos: %d pares de preferência`, j.JobID, j.Estado, j.DuracaoFormatada, j.Exemplos, j.Exemplos)
}

func formatarUsd(v float64) string {
	return strings.Replace(fmt.Sprintf("%.2f", v), ".", ",", 1)
}

func formatarMilhar(n int64) string {
	s := fmt.Sprintf("%d", n)
	var partes []string
	for len(s) > 3 {
		partes = append([]string{s[len(s)-3:]}, partes...)
		s = s[:len(s)-3]
	}
	partes = append([]string{s}, partes...)
	return strings.Join(partes, ".")
}

func GerarModelCardMarkdown(f FichaVersionamento) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Model Card, modelo fine-tunado\n\n")
	fmt.Fprintf(&sb, "## Identificação\n- Job: %s\n- Modelo ajustado: %s\n- Endpoint: %s\n- Estado: %s\n\n",
		f.JobID, f.ModeloAjustado, f.Endpoint, f.Estado)
	fmt.Fprintf(&sb, "## Linhagem\n- Modelo base: %s\n- Dataset de treino: %s\n- Hash SHA-256 do dataset: %s\n\n",
		f.ModeloBase, f.DatasetUri, f.DatasetHashSha256)
	epoch := int64(0)
	if f.EpochCount != nil {
		epoch = *f.EpochCount
	}
	lr := 0.0
	if f.LearningRateMultiplier != nil {
		lr = *f.LearningRateMultiplier
	}
	fmt.Fprintf(&sb, "## Hiperparâmetros\n- Épocas: %d\n- Taxa de aprendizado (multiplicador): %v\n- Rank do adaptador (LoRA): %s\n\n",
		epoch, lr, f.AdapterSize)
	exemplos := int64(0)
	if f.ExemplosDataset != nil {
		exemplos = *f.ExemplosDataset
	}
	tokens := int64(0)
	if f.TokensCobraveis != nil {
		tokens = *f.TokensCobraveis
	}
	fmt.Fprintf(&sb, "## Estatística do dataset\n- Exemplos de treino: %d\n- Tokens cobráveis no total: %d\n\n", exemplos, tokens)
	fmt.Fprintf(&sb, "## Linha do tempo\n- Criado em: %s\n- Concluído em: %s\n\n", f.CriadoEm, f.ConcluidoEm)

	if f.CustoEstimado != nil {
		c := f.CustoEstimado
		fmt.Fprintf(&sb, "## Custo real\n- Duração real do job: %s\n", c.DuracaoFormatada)
		notaFinal := "Nota: a Vertex AI cobra por token de treino, não por hora de GPU alugada; a faixa de GPU acima é referência de mercado pra comparar com o custo de rodar o mesmo tipo de treino (LoRA) em infraestrutura própria, não a fatura real deste job."
		if f.CustoReal != nil {
			fmt.Fprintf(&sb, "- **Custo real, conferido no billing do Google Cloud (28/08/2026)**: R$%s (%s tokens faturáveis × %d épocas = %s unidades cobradas, à taxa real de R$0,00002909/unidade apurada no relatório de billing por SKU de agosto/2026)\n",
				formatarUsd(f.CustoReal.CustoReais), formatarMilhar(tokens), epoch, formatarMilhar(f.CustoReal.Unidades))
			notaFinal = "Nota: a Vertex AI cobra por token de treino, não por hora de GPU alugada; a faixa de GPU acima é referência de mercado pra comparar com o custo de rodar o mesmo tipo de treino (LoRA) em infraestrutura própria - o valor real deste job específico é o R$" + formatarUsd(f.CustoReal.CustoReais) + " conferido no billing, acima."
		}
		fmt.Fprintf(&sb, "- Faixa GPU cloud consumer (US$ 0,40-0,80/hora, cheatsheet do Módulo 1.3): US$ %s-%s\n", formatarUsd(c.ConsumerMinUsd), formatarUsd(c.ConsumerMaxUsd))
		fmt.Fprintf(&sb, "- Faixa GPU cloud H100 (US$ 2,50-4,00/hora, cheatsheet do Módulo 1.3): US$ %s-%s\n", formatarUsd(c.H100MinUsd), formatarUsd(c.H100MaxUsd))
		fmt.Fprintf(&sb, "%s\n\n", notaFinal)
	}

	fmt.Fprintf(&sb, "## Nota de validade (ago/2026)\nEste model card documenta um job real, rodado com %s. O processo -- upload, hiperparâmetro, versionamento -- é o mesmo independente da versão exata do modelo-base. A Google aposenta versões do Gemini com aviso prévio (a família 2.5 tem retirement anunciado pra 16/out/2026); antes de treinar você mesmo, confira em [Vertex AI release notes](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/release-notes) quais modelos têm suporte a fine-tuning supervisionado no momento.",
		f.ModeloBase)
	return sb.String()
}
