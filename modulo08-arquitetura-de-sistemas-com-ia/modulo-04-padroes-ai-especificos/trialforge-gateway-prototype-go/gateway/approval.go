package gateway

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ApprovalGate (Módulo 4.4) pede aprovação humana antes de um rascunho virar
// oficial. Os originais em JS/Python precisam de um cuidado especial com
// stdin não-interativo (readline recriado a cada chamada, ou EOF prematuro —
// ver comentário no .js) porque o runtime deles fecha a fonte de entrada
// sozinho quando ela é lida aos poucos. Em Go, um único *bufio.Reader criado
// uma vez sobre o io.Reader (interativo ou tubo/pipe) e reutilizado a cada
// chamada não sofre desse problema — cada ReadString retoma de onde a
// anterior parou, então basta NÃO recriar o Reader a cada chamada. É a mesma
// disciplina descrita no .js, só que sem precisar detectar isTTY.
type ApprovalGate struct {
	reader *bufio.Reader
	writer io.Writer
}

// NewApprovalGate cria um ApprovalGate lendo de `entrada` e escrevendo
// prompts em `saida`.
func NewApprovalGate(entrada io.Reader, saida io.Writer) *ApprovalGate {
	return &ApprovalGate{reader: bufio.NewReader(entrada), writer: saida}
}

// PedirAprovacaoHumana mostra o rascunho e lê uma linha de resposta — "s"
// (case-insensitive, com espaços em volta ignorados) aprova, qualquer outra
// coisa (inclusive EOF/linha vazia) rejeita.
func (g *ApprovalGate) PedirAprovacaoHumana(rascunho string) (bool, error) {
	fmt.Fprintln(g.writer, "\n[Approval Gate] Rascunho aguardando aprovação antes de virar oficial:")
	fmt.Fprintln(g.writer, "   ", rascunho)
	fmt.Fprint(g.writer, "\nAprovar? (s/n) ")

	linha, err := g.reader.ReadString('\n')
	if err != nil && linha == "" {
		if err == io.EOF {
			return false, nil // sem mais entrada — trata como reprovado, não como erro fatal
		}
		return false, err
	}
	resposta := strings.TrimSpace(strings.ToLower(linha))
	return resposta == "s", nil
}
