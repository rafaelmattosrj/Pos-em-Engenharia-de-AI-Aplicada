package reactagent

// CasoTeste e um cenario de teste da ferramenta ExecutarBuscaClausula.
type CasoTeste struct {
	Tema        string
	Jurisdicao  string
	EsperaAchar bool
}

// CasosTesteFerramenta cobre os mesmos 6 cenarios de CASOS_TESTE_FERRAMENTA
// em react-agent-prototype.js / .py — os dois bugs reais encontrados
// testando contra o modelo de verdade na sessao original: a grafia
// "jurisdicicao" (testada a parte, em ExecutarBuscaClausula_test.go) e o
// tema formulado com "adolescente"/"pediátrico" em vez de "menor".
// Compartilhada entre main.go (roda em console, como no original) e os
// testes Go.
var CasosTesteFerramenta = []CasoTeste{
	{Tema: "Assentimento para menores de idade em estudos clínicos", Jurisdicao: "ANVISA", EsperaAchar: true},
	{Tema: "Consentimento de adolescentes em pesquisa", Jurisdicao: "ANVISA", EsperaAchar: true},
	{Tema: "Cuidados pediátricos em ensaio clínico", Jurisdicao: "ANVISA", EsperaAchar: true},
	{Tema: "Consentimento informado de população adulta", Jurisdicao: "ANVISA", EsperaAchar: false},
	{Tema: "Assentimento para menores de idade", Jurisdicao: "FDA", EsperaAchar: false}, // jurisdição errada
	{Tema: "Termo de Consentimento Livre e Esclarecido (TCLE)", Jurisdicao: "ANVISA", EsperaAchar: false},
}
