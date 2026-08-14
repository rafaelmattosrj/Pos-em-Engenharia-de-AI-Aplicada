// Package csvtool converte dados CSV para JSON — equivalente a
// CsvToJsonTool.java, usando encoding/csv da stdlib no lugar do Jackson CSV
// (ambos tratam corretamente campos entre aspas contendo vírgulas).
package csvtool

import (
	"encoding/csv"
	"encoding/json"
	"strings"
)

// Convert converte uma string CSV (com header na primeira linha) em uma
// string JSON representando um array de objetos — cada linha vira um objeto
// com os valores das colunas como strings, igual ao Map<String,String> por
// linha da versão Java.
func Convert(csvContent string) (string, error) {
	if strings.TrimSpace(csvContent) == "" {
		return "[]", nil
	}

	reader := csv.NewReader(strings.NewReader(strings.TrimSpace(csvContent)))
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}
	if len(records) == 0 {
		return "[]", nil
	}

	header := records[0]
	rows := make([]map[string]string, 0, len(records)-1)
	for _, record := range records[1:] {
		row := make(map[string]string, len(header))
		for i, col := range header {
			if i < len(record) {
				row[col] = record[i]
			}
		}
		rows = append(rows, row)
	}

	out, err := json.Marshal(rows)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
