package csvtool

import (
	"encoding/json"
	"testing"
)

func TestConvert_HeaderWithTwoRows(t *testing.T) {
	csv := "produto,quantidade,preco\nNotebook,10,2500.00\nMouse,50,45.00\n"

	out, err := Convert(csv)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	var rows []map[string]string
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatalf("saida nao e JSON valido: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("esperava 2 linhas, obteve %d", len(rows))
	}

	if rows[0]["produto"] != "Notebook" || rows[0]["quantidade"] != "10" || rows[0]["preco"] != "2500.00" {
		t.Errorf("primeira linha inesperada: %+v", rows[0])
	}
	if rows[1]["produto"] != "Mouse" || rows[1]["quantidade"] != "50" {
		t.Errorf("segunda linha inesperada: %+v", rows[1])
	}
}

func TestConvert_EmptyOrBlankReturnsEmptyArray(t *testing.T) {
	cases := []string{"", "   "}
	for _, c := range cases {
		out, err := Convert(c)
		if err != nil {
			t.Fatalf("erro inesperado para %q: %v", c, err)
		}
		if out != "[]" {
			t.Errorf("esperava '[]' para %q, obteve %q", c, out)
		}
	}
}

func TestConvert_QuotedValuesWithCommas(t *testing.T) {
	csv := "produto,descricao,preco\n" +
		`Notebook,"Notebook Dell, 16GB RAM",2500.00` + "\n" +
		`Teclado,"Teclado mecanico, RGB",250.00` + "\n"

	out, err := Convert(csv)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	var rows []map[string]string
	json.Unmarshal([]byte(out), &rows)
	if len(rows) != 2 {
		t.Fatalf("esperava 2 linhas, obteve %d", len(rows))
	}
	if rows[0]["descricao"] != "Notebook Dell, 16GB RAM" {
		t.Errorf("descricao inesperada: %q", rows[0]["descricao"])
	}
	if rows[1]["descricao"] != "Teclado mecanico, RGB" {
		t.Errorf("descricao inesperada: %q", rows[1]["descricao"])
	}
}
