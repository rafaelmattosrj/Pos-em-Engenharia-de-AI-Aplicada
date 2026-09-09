// Package rankadapter porta rank-adapter-comparison-tool.js (Modulo 4.3,
// companion, demo parte 2). Mesmo padrao de adaptercomparison: parsing puro e
// portavel; disparo de `python3 -m mlx_lm generate` depende de MLX +
// adaptadores locais.
package rankadapter

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	ModeloBase = "mlx-community/gemma-4-e2b-it-bf16"
	MaxTokens  = 80
)

const Prompt = "Extraia segurado, placa e valor do orçamento de oficina abaixo.\n\n" +
	"BOA VISTA REPAROS AUTOMOTIVOS CNPJ 21.098.765/0001-32 Rua dos Mecanicos 310 Segurado: " +
	"Ricardo Alves Monteiro Placa do veiculo: JBR-9021 Data do sinistro: 09/06/2026 Valor das " +
	"pecas: R$ 1.850,00 Valor da mao de obra: R$ 970,00 Valor total do orcamento: R$ 2.820,00"

var tokensPattern = regexp.MustCompile(`Generation: (\d+) tokens`)

func Gabarito() map[string]any {
	return map[string]any{"segurado": "Ricardo Alves Monteiro", "placa": "JBR-9021", "valor": float64(2820)}
}

// Adapters devolve rank -> caminho do diretorio de adaptador, mesma
// convencao de nomes do original.
func Adapters(baseDir string) map[string]string {
	return map[string]string{
		"rank 4":  baseDir + "/mlx-adapters-rank4",
		"rank 8":  baseDir + "/mlx-adapters",
		"rank 16": baseDir + "/mlx-adapters-rank16",
	}
}

func MontarArgumentos(adapterPath string) []string {
	args := []string{"-m", "mlx_lm", "generate", "--model", ModeloBase, "--prompt", Prompt, "--max-tokens", strconv.Itoa(MaxTokens)}
	if adapterPath != "" {
		args = append(args, "--adapter-path", adapterPath)
	}
	return args
}

type ResultadoParse struct {
	Texto  string
	Tokens int
	JSON   map[string]any
}

func ParsearSaida(textoSaida string) (ResultadoParse, error) {
	blocos := strings.Split(textoSaida, "==========")
	if len(blocos) < 3 {
		return ResultadoParse{}, fmt.Errorf("saída fora do formato esperado")
	}
	texto := strings.TrimSpace(blocos[1])
	tokens := 0
	if m := tokensPattern.FindStringSubmatch(textoSaida); m != nil {
		tokens, _ = strconv.Atoi(m[1])
	}
	var jsonResultado map[string]any
	_ = json.Unmarshal([]byte(texto), &jsonResultado)
	return ResultadoParse{Texto: texto, Tokens: tokens, JSON: jsonResultado}, nil
}

func BaterComGabarito(jsonResultado map[string]any) bool {
	if jsonResultado == nil {
		return false
	}
	for chave, esperado := range Gabarito() {
		obtido, existe := jsonResultado[chave]
		if !existe || fmt.Sprintf("%v", obtido) != fmt.Sprintf("%v", esperado) {
			return false
		}
	}
	return true
}

func RodarReal(adapterPath string) (string, error) {
	cmd := exec.Command("python3", MontarArgumentos(adapterPath)...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("mlx_lm generate falhou: %w:\n%s", err, stderr.String())
	}
	return stdout.String() + "\n" + stderr.String(), nil
}
