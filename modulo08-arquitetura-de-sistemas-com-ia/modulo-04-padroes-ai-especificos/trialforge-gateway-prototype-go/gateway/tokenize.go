package gateway

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// nonWordRe casa qualquer caractere que não seja "\w" (letra/dígito/underscore
// ASCII) nem espaço — mesmo conjunto que \w em JS sem a flag "u" e que \w do
// módulo `re` do Python quando o texto já foi reduzido a ASCII pelo
// stripAccents abaixo.
var nonWordRe = regexp.MustCompile(`[^\w\s]`)

// Tokenizar normaliza para minúsculas, decompõe acentos (NFD) e remove as
// marcas combinantes resultantes, troca pontuação por espaço e separa por
// espaço em branco. Mesma receita usada nos dois originais para preparar
// tanto o corpus quanto as queries do BM25 — idf mais estável com corpus
// pequeno quando os acentos não geram tokens distintos por coincidência.
func Tokenizar(texto string) []string {
	minusculo := strings.ToLower(texto)
	semAcento := removerAcentos(minusculo)
	semPontuacao := nonWordRe.ReplaceAllString(semAcento, " ")
	return strings.Fields(semPontuacao)
}

func removerAcentos(s string) string {
	decomposto := norm.NFD.String(s)
	var construtor strings.Builder
	construtor.Grow(len(decomposto))
	for _, r := range decomposto {
		if unicode.Is(unicode.Mn, r) {
			continue // marca combinante (acento) pós-NFD
		}
		construtor.WriteRune(r)
	}
	return construtor.String()
}
