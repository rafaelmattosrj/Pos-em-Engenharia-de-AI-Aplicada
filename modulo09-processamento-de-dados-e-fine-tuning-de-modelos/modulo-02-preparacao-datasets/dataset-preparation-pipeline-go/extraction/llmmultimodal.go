package extraction

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Porte de extracao-llm-multimodal-tool.js: extracao via LLM multimodal
// (Gemini/Vertex AI), em oposicao ao pipeline de OCR classico deste pacote.
//
// ObterTokenAcesso invoca o binario `gcloud` (equivalente ao execSync do
// original) e ExtrairViaLlm chama a API real via net/http - requer
// `gcloud auth application-default login` feito nesta maquina e rede, nao
// roda em CI/teste automatizado. CompararComEsperado e Normalizar sao
// puras e testadas isoladamente, igual ao original.

const (
	projetoGCP = "amplitude-seguros-demo"
	regiaoGCP  = "us-central1"
	modeloGCP  = "gemini-2.5-flash"
)

// ObterTokenAcesso obtem o token de acesso via `gcloud auth print-access-token`.
func ObterTokenAcesso() (string, error) {
	out, err := exec.Command("gcloud", "auth", "print-access-token").Output()
	if err != nil || len(bytes.TrimSpace(out)) == 0 {
		return "", fmt.Errorf(`nao consegui obter um token de acesso via gcloud. Rode "gcloud auth login" nesta maquina antes de rodar a demo ao vivo: %w`, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// removedorAcentos troca as vogais/consoantes acentuadas do portugues por sua
// forma sem acento. Evita dependencia externa (golang.org/x/text) so para
// stripping de diacritico - suficiente para o alfabeto usado neste dataset.
var removedorAcentos = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
)

var espacosMultiplos = regexp.MustCompile(`\s+`)

// Normalizar remove acento, baixa a caixa e colapsa espacos - usado para
// comparacao tolerante a variacao de OCR/LLM.
func Normalizar(valor any) string {
	s := fmt.Sprintf("%v", valor)
	s = strings.ToLower(strings.TrimSpace(s))
	s = removedorAcentos.Replace(s)
	return espacosMultiplos.ReplaceAllString(s, " ")
}

type geminiRequestPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inline_data,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiRequest struct {
	Contents []struct {
		Role  string              `json:"role"`
		Parts []geminiRequestPart `json:"parts"`
	} `json:"contents"`
	GenerationConfig struct {
		Temperature      float64 `json:"temperature"`
		ResponseMimeType string  `json:"responseMimeType"`
	} `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     *int64 `json:"promptTokenCount"`
		CandidatesTokenCount *int64 `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}

type ResultadoExtracao struct {
	Extraido      map[string]any
	LatenciaMs    int64
	TokensEntrada *int64
	TokensSaida   *int64
}

// ExtrairViaLlm chama o Gemini multimodal via Vertex AI com a imagem +
// instrucao, pedindo JSON estruturado sem passo de regex intermediario.
func ExtrairViaLlm(caminhoImagem, instrucao string, campos []string) (ResultadoExtracao, error) {
	bytesImagem, err := os.ReadFile(caminhoImagem)
	if err != nil {
		return ResultadoExtracao{}, err
	}
	imagemBase64 := base64.StdEncoding.EncodeToString(bytesImagem)
	listaCampos := strings.Join(campos, ", ")

	prompt := instrucao + " Devolva SOMENTE um JSON valido, sem markdown " +
		"e sem texto extra, com exatamente estas chaves: " + listaCampos + ". O campo \"valor\" deve " +
		"ser um numero (nao string, sem simbolo de moeda)."

	token, err := ObterTokenAcesso()
	if err != nil {
		return ResultadoExtracao{}, err
	}

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		regiaoGCP, projetoGCP, regiaoGCP, modeloGCP)

	var req geminiRequest
	req.Contents = []struct {
		Role  string              `json:"role"`
		Parts []geminiRequestPart `json:"parts"`
	}{{
		Role: "user",
		Parts: []geminiRequestPart{
			{Text: prompt},
			{InlineData: &inlineData{MimeType: "image/png", Data: imagemBase64}},
		},
	}}
	req.GenerationConfig.Temperature = 0
	req.GenerationConfig.ResponseMimeType = "application/json"

	corpo, err := json.Marshal(req)
	if err != nil {
		return ResultadoExtracao{}, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(corpo))
	if err != nil {
		return ResultadoExtracao{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	inicio := time.Now()
	resp, err := client.Do(httpReq)
	if err != nil {
		return ResultadoExtracao{}, err
	}
	defer resp.Body.Close()
	latenciaMs := time.Since(inicio).Milliseconds()

	corpoResposta, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResultadoExtracao{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ResultadoExtracao{}, fmt.Errorf("vertex AI retornou %d: %s", resp.StatusCode, string(corpoResposta))
	}

	var gResp geminiResponse
	if err := json.Unmarshal(corpoResposta, &gResp); err != nil {
		return ResultadoExtracao{}, err
	}
	if len(gResp.Candidates) == 0 || len(gResp.Candidates[0].Content.Parts) == 0 {
		return ResultadoExtracao{}, fmt.Errorf("resposta sem texto utilizavel: %s", string(corpoResposta))
	}
	textoResposta := gResp.Candidates[0].Content.Parts[0].Text

	var extraido map[string]any
	if err := json.Unmarshal([]byte(textoResposta), &extraido); err != nil {
		return ResultadoExtracao{}, fmt.Errorf("resposta nao e JSON valido: %s", textoResposta)
	}

	return ResultadoExtracao{
		Extraido: extraido, LatenciaMs: latenciaMs,
		TokensEntrada: gResp.UsageMetadata.PromptTokenCount, TokensSaida: gResp.UsageMetadata.CandidatesTokenCount,
	}, nil
}

type CampoComparado struct {
	Esperado any
	Extraido any
	Bate     bool
}

type ResultadoComparacao struct {
	PorCampo map[string]CampoComparado
	Acertos  int
	Total    int
}

// CompararComEsperado compara o extraido com o gabarito, campo a campo, com
// tolerancia a acento/caixa para strings e comparacao numerica para valores.
func CompararComEsperado(extraido, esperado map[string]any) ResultadoComparacao {
	porCampo := map[string]CampoComparado{}
	acertos := 0
	for campo, valorEsperado := range esperado {
		valorExtraido := extraido[campo]
		var bate bool
		if n, isNum := asFloat(valorEsperado); isNum {
			ne, okExtraido := asFloat(valorExtraido)
			bate = okExtraido && ne == n
		} else {
			ve := valorExtraido
			if ve == nil {
				ve = ""
			}
			bate = Normalizar(ve) == Normalizar(valorEsperado)
		}
		porCampo[campo] = CampoComparado{Esperado: valorEsperado, Extraido: valorExtraido, Bate: bate}
		if bate {
			acertos++
		}
	}
	return ResultadoComparacao{PorCampo: porCampo, Acertos: acertos, Total: len(esperado)}
}

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	default:
		return 0, false
	}
}
