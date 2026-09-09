package team

import (
	"fmt"
	"strings"

	"opspilot/domain"
	"opspilot/llm"
	"opspilot/tools"
)

// Porte de agents/roles.ts (team/roles.ts) -- os 3 papéis restritos
// (analista/planejador/executor).

const AnalistaSystemPrompt = "Voce e o ANALISTA do plantao. Sua unica funcao: produzir diagnostico FACTUAL do\n" +
	"estado atual, usando as ferramentas de leitura. NAO proponha solucoes. NAO abra\n" +
	"nem resolva nada."

const PlanejadorSystemPrompt = "Voce e o PLANEJADOR do plantao. Transforme os fatos do blackboard em um plano\n" +
	"numerado e executavel. Voce nao tem ferramentas: nao invente dados que nao estejam\n" +
	"no blackboard."

const ExecutorSystemPrompt = "Voce e o EXECUTOR do plantao. Execute acoes de incidente (abrir, resolver, listar)\n" +
	"conforme o brief e o plano do blackboard, usando somente as ferramentas disponiveis."

const defaultMaxIterations = 6

// RoleRunInput é o porte de RoleRunInput (team/roles.ts).
type RoleRunInput struct {
	Message    string
	Brief      string
	Blackboard []BlackboardEntry
}

// RoleRunResult é o porte de RoleRunResult (team/roles.ts).
type RoleRunResult struct {
	Entry    BlackboardEntry
	Trace    []domain.TraceEvent
	LLMCalls int
}

// RoleRunner é o porte da interface RoleRunner (team/roles.ts): cada papel é
// uma struct de agente com um método Run, orquestrada pela função
// coordenadora TeamGraph.Run.
type RoleRunner interface {
	Role() Role
	Tools() []tools.Tool
	Run(input RoleRunInput) (RoleRunResult, error)
}

func roleUserMessage(input RoleRunInput) string {
	return fmt.Sprintf(
		"Pedido original do plantonista: %s\nSua tarefa (brief do supervisor): %s\n\nBlackboard atual:\n%s",
		input.Message, input.Brief, RenderBlackboard(input.Blackboard))
}

// toolRoleRunner cobre analista e executor: ambos rodam um loop
// observação->ação com ferramentas, diferindo só em papel/kind/prompt/tools.
type toolRoleRunner struct {
	role          Role
	kind          BlackboardKind
	prompt        string
	model         llm.ChatModel
	toolset       []tools.Tool
	maxIterations int
}

// NewAnalistaRunner é o porte de createAnalistaRunner.
func NewAnalistaRunner(model llm.ChatModel, toolset []tools.Tool) RoleRunner {
	return &toolRoleRunner{
		role: RoleAnalista, kind: KindFacts, prompt: AnalistaSystemPrompt,
		model: model, toolset: toolset, maxIterations: defaultMaxIterations,
	}
}

// NewExecutorRunner é o porte de createExecutorRunner.
func NewExecutorRunner(model llm.ChatModel, toolset []tools.Tool) RoleRunner {
	return &toolRoleRunner{
		role: RoleExecutor, kind: KindExecution, prompt: ExecutorSystemPrompt,
		model: model, toolset: toolset, maxIterations: defaultMaxIterations,
	}
}

func (r *toolRoleRunner) Role() Role          { return r.role }
func (r *toolRoleRunner) Tools() []tools.Tool { return r.toolset }

func (r *toolRoleRunner) Run(input RoleRunInput) (RoleRunResult, error) {
	messages := []llm.ChatMessage{
		llm.SystemMessage(r.prompt),
		llm.UserMessage(roleUserMessage(input)),
	}
	var trace []domain.TraceEvent
	llmCalls := 0
	finalAnswer := ""
	nodeName := strings.ToLower(string(r.role))

	for iteration := 0; iteration < r.maxIterations; iteration++ {
		response, err := r.model.Invoke(messages, r.toolset)
		llmCalls++
		if err != nil {
			return RoleRunResult{}, err
		}

		if len(response.ToolCalls) == 0 {
			finalAnswer = response.Content
			trace = append(trace, domain.Answer(nodeName, finalAnswer))
			break
		}

		for _, call := range response.ToolCalls {
			trace = append(trace, domain.Action(nodeName, fmt.Sprintf("%s %v", call.ToolName, call.Args)))
			observation := executeToolFor(r.toolset, call)
			trace = append(trace, domain.Observation(nodeName, observation))
			messages = append(messages, llm.ToolMessage(observation))
		}
	}

	entry := BlackboardEntry{Role: r.role, Kind: r.kind, Brief: input.Brief, Content: finalAnswer}
	return RoleRunResult{Entry: entry, Trace: trace, LLMCalls: llmCalls}, nil
}

// planejadorRunner: zero ferramentas por assinatura -- só uma chamada de
// modelo sobre o blackboard.
type planejadorRunner struct {
	model llm.ChatModel
}

// NewPlanejadorRunner é o porte de createPlanejadorRunner.
func NewPlanejadorRunner(model llm.ChatModel) RoleRunner {
	return &planejadorRunner{model: model}
}

func (r *planejadorRunner) Role() Role          { return RolePlanejador }
func (r *planejadorRunner) Tools() []tools.Tool { return nil }

func (r *planejadorRunner) Run(input RoleRunInput) (RoleRunResult, error) {
	response, err := r.model.Invoke([]llm.ChatMessage{
		llm.SystemMessage(PlanejadorSystemPrompt),
		llm.UserMessage(roleUserMessage(input)),
	}, nil)
	if err != nil {
		return RoleRunResult{}, err
	}

	entry := BlackboardEntry{Role: RolePlanejador, Kind: KindPlan, Brief: input.Brief, Content: response.Content}
	return RoleRunResult{
		Entry:    entry,
		Trace:    []domain.TraceEvent{domain.Plan("planejador", response.Content)},
		LLMCalls: 1,
	}, nil
}

func executeToolFor(toolset []tools.Tool, call llm.ToolCall) string {
	for _, tool := range toolset {
		if tool.Name() == call.ToolName {
			result, err := tool.Execute(call.Args)
			if err != nil {
				return "Erro: " + err.Error()
			}
			return result
		}
	}
	return fmt.Sprintf("Erro: ferramenta %q nao disponivel para este papel.", call.ToolName)
}
