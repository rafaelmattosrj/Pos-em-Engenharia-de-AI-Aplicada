// Package splitter divide um texto longo em chunks menores com overlap —
// equivalente a DocumentSplitters.recursive(chunkSize, overlap) do
// LangChain4j. A estratégia é recursiva: o texto é primeiro quebrado em
// unidades pequenas (parágrafo → linha → frase → palavra, na ordem dos
// separadores abaixo), que depois são empacotadas em chunks de até
// chunkSize caracteres, sempre preservando os últimos `overlap` caracteres
// do chunk anterior no início do próximo.
package splitter

import "strings"

var separators = []string{"\n\n", "\n", ". ", " "}

// Recursive divide text em chunks de até chunkSize caracteres, com overlap
// caracteres repetidos entre chunks consecutivos.
func Recursive(text string, chunkSize, overlap int) []string {
	units := descend(text, 0)
	return pack(units, chunkSize, overlap)
}

// descend quebra text recursivamente pelos separadores, do mais "grosso"
// (parágrafo) ao mais "fino" (palavra), retornando as menores unidades
// possíveis. Sempre termina: level cresce a cada chamada recursiva.
func descend(text string, level int) []string {
	if level >= len(separators) {
		return []string{text}
	}

	parts := splitKeepSeparator(text, separators[level])
	if len(parts) <= 1 {
		return descend(text, level+1)
	}

	var out []string
	for _, p := range parts {
		out = append(out, descend(p, level+1)...)
	}
	return out
}

// splitKeepSeparator quebra text em sep, reanexando o separador removido a
// cada pedaço (exceto o último) para preservar espaços/quebras de linha.
func splitKeepSeparator(text, sep string) []string {
	parts := strings.Split(text, sep)
	var out []string
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i < len(parts)-1 {
			part += sep
		}
		out = append(out, part)
	}
	return out
}

// pack empacota as unidades em chunks de até chunkSize caracteres. Uma
// unidade sozinha maior que chunkSize (ex.: texto sem nenhum separador) é
// dividida em pedaços de tamanho fixo.
func pack(units []string, chunkSize, overlap int) []string {
	var chunks []string
	var current strings.Builder

	flush := func() {
		if current.Len() == 0 {
			return
		}
		content := current.String()
		chunks = append(chunks, content)
		current.Reset()
		if overlap > 0 && len(content) > overlap {
			current.WriteString(content[len(content)-overlap:])
		}
	}

	addUnit := func(unit string) {
		if current.Len() > 0 && current.Len()+len(unit) > chunkSize {
			flush()
		}
		current.WriteString(unit)
	}

	for _, u := range units {
		if len(u) > chunkSize {
			for _, piece := range hardSplit(u, chunkSize) {
				addUnit(piece)
			}
			continue
		}
		addUnit(u)
	}
	flush()

	return chunks
}

// hardSplit divide text em pedaços de exatamente `size` runas, usado como
// último recurso quando não há separador algum para uma unidade grande
// demais.
func hardSplit(text string, size int) []string {
	runes := []rune(text)
	var out []string
	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[i:end]))
	}
	return out
}
