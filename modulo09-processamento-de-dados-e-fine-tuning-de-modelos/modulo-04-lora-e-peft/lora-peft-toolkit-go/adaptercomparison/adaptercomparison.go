// Package adaptercomparison porta adapter-comparison-tool.js (Modulo 4.2,
// companion). Parsing e montagem de argumentos sao logica pura e portavel;
// RodarGenerateReal dispara `python3 -m mlx_lm generate`, que so funciona com
// MLX + modelo/adaptador presentes localmente -- mesma dependencia de
// hardware do script original. O disparo de processo via os/exec e
// portavel; a inferencia MLX em si nao tem equivalente de biblioteca Go.
package adaptercomparison

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	ModeloBase    = "mlx-community/gemma-4-e2b-it-bf16"
	MaxTokens     = 80
	IndiceExemplo = 8
)

var tokensPattern = regexp.MustCompile(`Generation: (\d+) tokens`)

type ExemploTeste struct {
	Prompt   string
	Gabarito map[string]any
}

type mensagem struct {
	Content string `json:"content"`
}

type exemploJSONL struct {
	Messages []mensagem `json:"messages"`
}

// CarregarExemploTeste le uma linha (indice) de um arquivo JSONL de teste
// (mesmo formato de mlx-data/test.jsonl) e devolve o prompt + gabarito.
func CarregarExemploTeste(caminho string, indice int) (ExemploTeste, error) {
	arquivo, err := os.Open(caminho)
	if err != nil {
		return ExemploTeste{}, err
	}
	defer arquivo.Close()

	var linhas []string
	scanner := bufio.NewScanner(arquivo)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())
		if linha != "" {
			linhas = append(linhas, linha)
		}
	}
	if err := scanner.Err(); err != nil {
		return ExemploTeste{}, err
	}

	if indice < 0 || indice >= len(linhas) {
		return ExemploTeste{}, fmt.Errorf("índice %d fora do intervalo (0-%d)", indice, len(linhas)-1)
	}

	var exemplo exemploJSONL
	if err := json.Unmarshal([]byte(linhas[indice]), &exemplo); err != nil {
		return ExemploTeste{}, err
	}

	var gabarito map[string]any
	if err := json.Unmarshal([]byte(exemplo.Messages[1].Content), &gabarito); err != nil {
		return ExemploTeste{}, err
	}

	return ExemploTeste{Prompt: exemplo.Messages[0].Content, Gabarito: gabarito}, nil
}

func MontarArgumentosGenerate(prompt, adapterPath string, maxTokens int, modelo string) []string {
	args := []string{"-m", "mlx_lm", "generate", "--model", modelo, "--prompt", prompt, "--max-tokens", strconv.Itoa(maxTokens)}
	if adapterPath != "" {
		args = append(args, "--adapter-path", adapterPath)
	}
	return args
}

type ResultadoParse struct {
	TextoGerado           string
	TokensGerados         int
	TokensConhecidos      bool
	JSON                  map[string]any
	BateuNoLimiteDeTokens bool
}

func ParsearSaidaGenerate(textoSaida string) (ResultadoParse, error) {
	blocos := strings.Split(textoSaida, "==========")
	if len(blocos) < 3 {
		return ResultadoParse{}, fmt.Errorf("saída não tem o formato esperado (dois separadores \"==========\")")
	}
	textoGerado := strings.TrimSpace(blocos[1])

	var tokensGerados int
	tokensConhecidos := false
	if m := tokensPattern.FindStringSubmatch(textoSaida); m != nil {
		tokensGerados, _ = strconv.Atoi(m[1])
		tokensConhecidos = true
	}

	var jsonResultado map[string]any
	_ = json.Unmarshal([]byte(textoGerado), &jsonResultado) // nao-JSON -> jsonResultado fica nil, igual ao original

	bateuNoLimite := tokensConhecidos && tokensGerados >= MaxTokens

	return ResultadoParse{
		TextoGerado: textoGerado, TokensGerados: tokensGerados, TokensConhecidos: tokensConhecidos,
		JSON: jsonResultado, BateuNoLimiteDeTokens: bateuNoLimite,
	}, nil
}

func CompararComGabarito(jsonResultado, gabarito map[string]any) bool {
	if jsonResultado == nil {
		return false
	}
	for chave, valorEsperado := range gabarito {
		valorObtido, existe := jsonResultado[chave]
		if !existe || fmt.Sprintf("%v", valorObtido) != fmt.Sprintf("%v", valorEsperado) {
			return false
		}
	}
	return true
}

// RodarGenerateReal dispara `python3 <args...>` e devolve stdout+stderr
// combinados (mesmo comportamento do spawnSync original, que tambem
// concatena os dois fluxos).
func RodarGenerateReal(args []string) (string, error) {
	cmd := exec.Command("python3", args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("mlx_lm generate falhou: %w:\n%s", err, stderr.String())
	}
	return stdout.String() + "\n" + stderr.String(), nil
}
