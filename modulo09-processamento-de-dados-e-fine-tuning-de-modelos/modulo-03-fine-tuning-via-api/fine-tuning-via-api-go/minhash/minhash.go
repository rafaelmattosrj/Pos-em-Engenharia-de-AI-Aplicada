// Package minhash implementa deduplicacao via MinHash+LSH e balanceamento
// por amostragem com temperatura + metricas de diversidade (entropia de
// Shannon / numero efetivo de fontes).
//
// ADAPTACAO: este e' um porte AUTOCONTIDO do algoritmo de
// modulo-02-preparacao-datasets/dataset-cleaning-balancing-tool.js (mesma
// formula, mesmos parametros MinHashK=32/semente=42/LSH bandas=8,linhas=4,
// alpha=0.3, limiar de duplicata=0.55). O original em JS reusa o mesmo
// arquivo via require() entre modulo-02 e modulo-03/09; como este repo nao
// tem um mecanismo de modulo compartilhado entre projetos Maven/Go
// independentes, a logica foi duplicada aqui deliberadamente (nao portada
// por referencia), documentado tambem no README deste projeto.
package minhash

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

const (
	MinHashK         = 32
	MinHashSemente   = 42
	LSHBandas        = 8
	LSHLinhas        = 4
	LimiarDuplicata  = 0.55
	AlphaTemperatura = 0.3
)

const primoMersenne = 2147483647 // 2^31 - 1

var espacos = regexp.MustCompile(`\s+`)

func NormalizarTexto(texto string) string {
	return strings.TrimSpace(espacos.ReplaceAllString(strings.ToLower(texto), " "))
}

func Shingles(texto string, n int) map[string]bool {
	palavras := strings.Split(NormalizarTexto(texto), " ")
	conjunto := make(map[string]bool)
	for i := 0; i <= len(palavras)-n; i++ {
		conjunto[strings.Join(palavras[i:i+n], " ")] = true
	}
	return conjunto
}

func SimilaridadeJaccardExata(a, b string, n int) float64 {
	sa := Shingles(a, n)
	sb := Shingles(b, n)
	if len(sa) == 0 || len(sb) == 0 {
		return 0
	}
	intersecao := 0
	for s := range sa {
		if sb[s] {
			intersecao++
		}
	}
	uniao := len(sa) + len(sb) - intersecao
	if uniao == 0 {
		return 0
	}
	return float64(intersecao) / float64(uniao)
}

// hashString itera por rune (nao por byte) pra bater com o s.charCodeAt(i) do
// JS original: texto real deste dataset tem acento (Clínica, São Rafael,
// Saúde), que em UTF-8 ocupa mais de 1 byte -- iterar por byte quebraria a
// paridade do hash com a implementacao JS.
func hashString(s string) uint32 {
	var h uint32 = 5381
	for _, r := range s {
		h = (h*33 + uint32(r))
	}
	return h
}

type CoeficienteHash struct{ A, B int64 }

func GerarCoeficientesHash(k int, semente int64) []CoeficienteHash {
	estado := uint32(semente)
	proximo := func() uint32 {
		estado = uint32(uint64(estado)*1103515245+12345) & 0xffffffff
		return estado
	}
	coeficientes := make([]CoeficienteHash, k)
	for i := 0; i < k; i++ {
		a := int64(proximo()%2000000000) + 1
		b := int64(proximo() % 2000000000)
		coeficientes[i] = CoeficienteHash{A: a, B: b}
	}
	return coeficientes
}

func hashUniversal(x, a, b int64) int64 {
	return (a*x + b) % primoMersenne
}

func AssinaturaMinHash(shingleSet map[string]bool, coeficientes []CoeficienteHash) []int64 {
	baseHashes := make([]int64, 0, len(shingleSet))
	for s := range shingleSet {
		baseHashes = append(baseHashes, int64(hashString(s)))
	}
	assinatura := make([]int64, len(coeficientes))
	for i, c := range coeficientes {
		minimo := int64(math.MaxInt64)
		for _, x := range baseHashes {
			h := hashUniversal(x, c.A, c.B)
			if h < minimo {
				minimo = h
			}
		}
		assinatura[i] = minimo
	}
	return assinatura
}

func SimilaridadeMinHashEstimada(a, b []int64) float64 {
	iguais := 0
	for i := range a {
		if a[i] == b[i] {
			iguais++
		}
	}
	return float64(iguais) / float64(len(a))
}

func BandingLSH(assinaturas [][]int64, b, r int) map[string]bool {
	baldes := make(map[string][]int)
	candidatos := make(map[string]bool)
	for idx, assinatura := range assinaturas {
		for banda := 0; banda < b; banda++ {
			var fatia strings.Builder
			for k := banda * r; k < banda*r+r; k++ {
				fmt.Fprintf(&fatia, "%d,", assinatura[k])
			}
			chave := fmt.Sprintf("%d:%s", banda, fatia.String())
			for _, outroIdx := range baldes[chave] {
				if outroIdx < idx {
					candidatos[fmt.Sprintf("%d-%d", outroIdx, idx)] = true
				} else {
					candidatos[fmt.Sprintf("%d-%d", idx, outroIdx)] = true
				}
			}
			baldes[chave] = append(baldes[chave], idx)
		}
	}
	return candidatos
}

type Exemplo struct {
	ID, Caso, Fonte, TextoParaDedup string
}

type ParDuplicata struct {
	I, J          int
	Similaridade float64
}

type ResultadoDedup struct {
	ParesDuplicata                       []ParDuplicata
	TotalParesForcaBruta, TotalCandidatosLSH int
}

// EncontrarQuaseDuplicatasGenerico roda dedup restrito a um conjunto de
// exemplos generico -- nao hardcoded pra casos especificos.
func EncontrarQuaseDuplicatasGenerico(exemplos []Exemplo, nShingle int) ResultadoDedup {
	coeficientes := GerarCoeficientesHash(MinHashK, MinHashSemente)
	assinaturas := make([][]int64, len(exemplos))
	for i, e := range exemplos {
		assinaturas[i] = AssinaturaMinHash(Shingles(e.TextoParaDedup, nShingle), coeficientes)
	}
	candidatos := BandingLSH(assinaturas, LSHBandas, LSHLinhas)
	totalForcaBruta := len(exemplos) * (len(exemplos) - 1) / 2

	var pares []ParDuplicata
	for chave := range candidatos {
		var li, lj int
		fmt.Sscanf(chave, "%d-%d", &li, &lj)
		sim := SimilaridadeJaccardExata(exemplos[li].TextoParaDedup, exemplos[lj].TextoParaDedup, nShingle)
		if sim >= LimiarDuplicata {
			pares = append(pares, ParDuplicata{I: li, J: lj, Similaridade: sim})
		}
	}
	return ResultadoDedup{ParesDuplicata: pares, TotalParesForcaBruta: totalForcaBruta, TotalCandidatosLSH: len(candidatos)}
}

func ContarPorFonte(exemplos []Exemplo, caso string) map[string]int {
	contagem := make(map[string]int)
	for _, e := range exemplos {
		if e.Caso == caso {
			contagem[e.Fonte]++
		}
	}
	return contagem
}

func PesosAmostragemPorTemperatura(contagens map[string]int, alpha float64) map[string]float64 {
	fontes := chavesOrdenadas(contagens)
	pesosBrutos := make(map[string]float64, len(fontes))
	soma := 0.0
	for _, f := range fontes {
		p := math.Pow(float64(contagens[f]), alpha)
		pesosBrutos[f] = p
		soma += p
	}
	pesos := make(map[string]float64, len(fontes))
	for _, f := range fontes {
		pesos[f] = pesosBrutos[f] / soma
	}
	return pesos
}

func chavesOrdenadas(m map[string]int) []string {
	chaves := make([]string, 0, len(m))
	for k := range m {
		chaves = append(chaves, k)
	}
	sort.Strings(chaves)
	return chaves
}

func AlocarMaiorResto(pesos map[string]float64, alvo int) map[string]int {
	fontes := make([]string, 0, len(pesos))
	for f := range pesos {
		fontes = append(fontes, f)
	}
	sort.Strings(fontes)

	quotas := make([]float64, len(fontes))
	base := make([]int, len(fontes))
	alocadoBase := 0
	for i, f := range fontes {
		quotas[i] = pesos[f] * float64(alvo)
		base[i] = int(math.Floor(quotas[i]))
		alocadoBase += base[i]
	}
	ordem := make([]int, len(fontes))
	for i := range ordem {
		ordem[i] = i
	}
	sort.Slice(ordem, func(x, y int) bool {
		return (quotas[ordem[y]] - float64(base[ordem[y]])) < (quotas[ordem[x]] - float64(base[ordem[x]]))
	})

	resultado := make(map[string]int, len(fontes))
	for i, f := range fontes {
		resultado[f] = base[i]
	}
	faltam := alvo - alocadoBase
	for i := 0; i < faltam; i++ {
		resultado[fontes[ordem[i]]]++
	}
	return resultado
}

// AlocarComCapacidade nunca aloca mais do que a fonte realmente tem disponivel.
func AlocarComCapacidade(contagens map[string]int, alpha float64, alvoTotal int) map[string]int {
	fontesAtivas := chavesOrdenadas(contagens)
	alvoRestante := alvoTotal
	resultado := make(map[string]int)
	maxIteracoes := len(contagens) + 1

	for iter := 0; iter < maxIteracoes && len(fontesAtivas) > 0; iter++ {
		contagensAtivas := make(map[string]int, len(fontesAtivas))
		for _, f := range fontesAtivas {
			contagensAtivas[f] = contagens[f]
		}
		pesos := PesosAmostragemPorTemperatura(contagensAtivas, alpha)
		tentativa := AlocarMaiorResto(pesos, alvoRestante)

		var excedentes []string
		for _, f := range fontesAtivas {
			if tentativa[f] > contagens[f] {
				excedentes = append(excedentes, f)
			}
		}
		if len(excedentes) == 0 {
			for _, f := range fontesAtivas {
				resultado[f] = tentativa[f]
			}
			break
		}
		excedenteSet := make(map[string]bool, len(excedentes))
		for _, f := range excedentes {
			resultado[f] = contagens[f]
			alvoRestante -= contagens[f]
			excedenteSet[f] = true
		}
		var restantes []string
		for _, f := range fontesAtivas {
			if !excedenteSet[f] {
				restantes = append(restantes, f)
			}
		}
		fontesAtivas = restantes
	}
	return resultado
}

func DistribuicaoDe(contagens map[string]int) map[string]float64 {
	total := 0
	for _, n := range contagens {
		total += n
	}
	dist := make(map[string]float64, len(contagens))
	for f, n := range contagens {
		if total == 0 {
			dist[f] = 0
		} else {
			dist[f] = float64(n) / float64(total)
		}
	}
	return dist
}

func EntropiaShannon(distribuicao map[string]float64) float64 {
	soma := 0.0
	for _, p := range distribuicao {
		if p > 0 {
			soma += p * math.Log(p)
		}
	}
	return -soma
}

func NumeroEfetivoFontes(distribuicao map[string]float64) float64 {
	return math.Exp(EntropiaShannon(distribuicao))
}
