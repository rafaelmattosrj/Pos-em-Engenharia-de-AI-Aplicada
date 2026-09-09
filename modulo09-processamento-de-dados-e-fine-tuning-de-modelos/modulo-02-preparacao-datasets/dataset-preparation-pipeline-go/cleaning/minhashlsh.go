package cleaning

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	LimiarDuplicata = 0.55
	MinhashK        = 32
	MinhashSemente  = 42
	LshBandas       = 8
	LshLinhas       = 4
	primoMersenne   = 2147483647 // 2^31 - 1
)

var espacos = regexp.MustCompile(`\s+`)

func NormalizarTexto(texto string) string {
	return strings.TrimSpace(espacos.ReplaceAllString(strings.ToLower(texto), " "))
}

// Shingles gera os n-gramas de palavras (shingles) do texto normalizado.
func Shingles(texto string, n int) map[string]struct{} {
	palavras := strings.Split(NormalizarTexto(texto), " ")
	conjunto := map[string]struct{}{}
	for i := 0; i <= len(palavras)-n; i++ {
		conjunto[strings.Join(palavras[i:i+n], " ")] = struct{}{}
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
		if _, ok := sb[s]; ok {
			intersecao++
		}
	}
	uniao := len(sa) + len(sb) - intersecao
	if uniao == 0 {
		return 0
	}
	return float64(intersecao) / float64(uniao)
}

// hashString: djb2, hash de string deterministico, 32 bits sem sinal.
func hashString(s string) uint32 {
	var h uint32 = 5381
	for i := 0; i < len(s); i++ {
		h = (h*33 + uint32(s[i])) & 0xffffffff
	}
	return h
}

type CoeficienteHash struct {
	A, B int64
}

// GerarCoeficientesHash: LCG deterministico (mesma semente sempre produz a
// mesma familia de funcoes hash - necessario pra reprodutibilidade).
func GerarCoeficientesHash(k int, semente int64) []CoeficienteHash {
	estado := semente & 0xffffffff
	proximo := func() int64 {
		estado = (estado*1103515245 + 12345) & 0xffffffff
		return estado
	}
	coeficientes := make([]CoeficienteHash, 0, k)
	for i := 0; i < k; i++ {
		a := proximo()%2000000000 + 1
		b := proximo() % 2000000000
		coeficientes = append(coeficientes, CoeficienteHash{A: a, B: b})
	}
	return coeficientes
}

func hashUniversal(x, a, b int64) int64 {
	return (a*x + b) % primoMersenne
}

// AssinaturaMinHash: k valores minimos, um por funcao hash, sobre o conjunto de shingles.
func AssinaturaMinHash(shingleSet map[string]struct{}, coeficientes []CoeficienteHash) []int64 {
	baseHashes := make([]int64, 0, len(shingleSet))
	for s := range shingleSet {
		baseHashes = append(baseHashes, int64(hashString(s)))
	}
	assinatura := make([]int64, len(coeficientes))
	for i, c := range coeficientes {
		minimo := int64(1) << 62
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

// BandingLSH: divide a assinatura de k valores em b bandas de r valores cada
// (k = b*r). Dois exemplos viram "candidatos" se colidirem em pelo menos uma banda.
func BandingLSH(assinaturas [][]int64, b, r int) map[string]struct{} {
	baldes := map[string][]int{}
	candidatos := map[string]struct{}{}
	for idx, assinatura := range assinaturas {
		for banda := 0; banda < b; banda++ {
			partes := make([]string, r)
			for j := 0; j < r; j++ {
				partes[j] = strconv.FormatInt(assinatura[banda*r+j], 10)
			}
			chave := fmt.Sprintf("%d:%s", banda, strings.Join(partes, ","))
			for _, outroIdx := range baldes[chave] {
				menor, maior := outroIdx, idx
				if idx < outroIdx {
					menor, maior = idx, outroIdx
				}
				candidatos[fmt.Sprintf("%d-%d", menor, maior)] = struct{}{}
			}
			baldes[chave] = append(baldes[chave], idx)
		}
	}
	return candidatos
}

type ParDuplicata struct {
	I, J         int
	Similaridade float64
}

type ResultadoPorCaso struct {
	ItensNoCaso           int
	ParesForcaBruta       int64
	CandidatosLSH         int
	DuplicatasConfirmadas int
	ReducaoPercentual     float64
}

type ResultadoDedup struct {
	ParesDuplicata       []ParDuplicata
	ResultadosPorCaso    map[string]ResultadoPorCaso
	TotalParesForcaBruta int64
	TotalCandidatosLSH   int64
}

func EncontrarQuaseDuplicatasMinHashLSH(exemplos []Exemplo) ResultadoDedup {
	coeficientes := GerarCoeficientesHash(MinhashK, MinhashSemente)
	resultadosPorCaso := map[string]ResultadoPorCaso{}
	var totalParesForcaBruta, totalCandidatosLSH int64
	var paresDuplicata []ParDuplicata

	for _, caso := range []string{"amplitude-auto", "amplitude-saude-empresarial"} {
		var indicesGlobais []int
		var itens []Exemplo
		for i, e := range exemplos {
			if e.Metadata.Caso == caso {
				indicesGlobais = append(indicesGlobais, i)
				itens = append(itens, e)
			}
		}
		assinaturas := make([][]int64, len(itens))
		for i, e := range itens {
			assinaturas[i] = AssinaturaMinHash(Shingles(e.Entrada, 5), coeficientes)
		}
		candidatosLocais := BandingLSH(assinaturas, LshBandas, LshLinhas)
		paresForcaBruta := int64(len(itens)) * int64(len(itens)-1) / 2

		confirmados := 0
		for chave := range candidatosLocais {
			var li, lj int
			fmt.Sscanf(chave, "%d-%d", &li, &lj)
			simExata := SimilaridadeJaccardExata(itens[li].Entrada, itens[lj].Entrada, 5)
			if simExata >= LimiarDuplicata {
				confirmados++
				paresDuplicata = append(paresDuplicata, ParDuplicata{I: indicesGlobais[li], J: indicesGlobais[lj], Similaridade: simExata})
			}
		}

		reducaoPercentual := 0.0
		if paresForcaBruta != 0 {
			reducaoPercentual = 100 * (1 - float64(len(candidatosLocais))/float64(paresForcaBruta))
		}
		resultadosPorCaso[caso] = ResultadoPorCaso{
			ItensNoCaso: len(itens), ParesForcaBruta: paresForcaBruta, CandidatosLSH: len(candidatosLocais),
			DuplicatasConfirmadas: confirmados, ReducaoPercentual: reducaoPercentual,
		}
		totalParesForcaBruta += paresForcaBruta
		totalCandidatosLSH += int64(len(candidatosLocais))
	}

	return ResultadoDedup{
		ParesDuplicata: paresDuplicata, ResultadosPorCaso: resultadosPorCaso,
		TotalParesForcaBruta: totalParesForcaBruta, TotalCandidatosLSH: totalCandidatosLSH,
	}
}

type ResultadoRemocao struct {
	Mantidos             []Exemplo
	ParesDuplicata       []ParDuplicata
	ResultadosPorCaso    map[string]ResultadoPorCaso
	TotalParesForcaBruta int64
	TotalCandidatosLSH   int64
	Removidos            int
}

// RemoverQuaseDuplicatas mantem a primeira ocorrencia de cada par de
// quase-duplicata e remove a segunda.
func RemoverQuaseDuplicatas(exemplos []Exemplo) ResultadoRemocao {
	dedup := EncontrarQuaseDuplicatasMinHashLSH(exemplos)
	remover := map[int]struct{}{}
	for _, p := range dedup.ParesDuplicata {
		remover[p.J] = struct{}{}
	}
	var mantidos []Exemplo
	for i, e := range exemplos {
		if _, ok := remover[i]; !ok {
			mantidos = append(mantidos, e)
		}
	}
	return ResultadoRemocao{
		Mantidos: mantidos, ParesDuplicata: dedup.ParesDuplicata, ResultadosPorCaso: dedup.ResultadosPorCaso,
		TotalParesForcaBruta: dedup.TotalParesForcaBruta, TotalCandidatosLSH: dedup.TotalCandidatosLSH,
		Removidos: len(remover),
	}
}
