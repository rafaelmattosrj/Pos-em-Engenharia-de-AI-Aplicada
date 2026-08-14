package agente

import "strings"

// ResultadoClausula e o retorno de BuscarClausulaAssentimento: ou tem
// Texto+Fonte (clausula encontrada), ou tem Aviso (nao se aplica).
type ResultadoClausula struct {
	Texto string
	Fonte string
	Aviso string
}

// BuscarClausulaAssentimento e a versao minima de ferramenta deterministica
// que o agente aciona — so pra tornar tangivel "ferramenta = funcao
// deterministica que o agente delega". O schema formal de ferramenta (JSON,
// enum, validacao de parametro) fica a cargo de react-agent-prototype.
func BuscarClausulaAssentimento(faixaEtaria []int) ResultadoClausula {
	cobreMenor := false
	for _, idade := range faixaEtaria {
		if idade < 18 {
			cobreMenor = true
			break
		}
	}
	if !cobreMenor {
		return ResultadoClausula{Aviso: "população adulta: cláusula de assentimento não se aplica"}
	}
	return ResultadoClausula{
		Texto: "Para participantes entre 12 e 17 anos, é necessário assentimento por escrito, além do " +
			"consentimento do responsável legal (RDC ANVISA 466/2012, Art. 4º).",
		Fonte: "RDC ANVISA 466/2012, Art. 4º",
	}
}

// String formata o resultado como um mapa Go legivel, para os prints da demo
// (equivalente ao console.log de um objeto no original).
func (r ResultadoClausula) String() string {
	var b strings.Builder
	b.WriteString("{")
	if r.Texto != "" {
		b.WriteString("texto:\"" + r.Texto + "\" fonte:\"" + r.Fonte + "\"")
	} else {
		b.WriteString("texto:null aviso:\"" + r.Aviso + "\"")
	}
	b.WriteString("}")
	return b.String()
}
