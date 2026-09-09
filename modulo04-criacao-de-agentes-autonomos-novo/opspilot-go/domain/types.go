// Package domain contém os tipos centrais do núcleo portado do OpsPilot
// (domain/types.ts, domain/severity.ts): alertas, incidentes, runbooks,
// eventos de trace e o contrato ReasoningStrategy usado tanto pela estratégia
// ReAct de referência quanto pela estratégia multiagente (team).
package domain

// Severity é o porte de Severity (domain/severity.ts).
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
)

// AlertStatus é o porte do enum AlertStatus (domain/types.ts).
type AlertStatus string

const (
	AlertFiring   AlertStatus = "FIRING"
	AlertResolved AlertStatus = "RESOLVED"
)

// IncidentStatus é o porte do enum IncidentStatus (domain/types.ts).
type IncidentStatus string

const (
	IncidentOpen     IncidentStatus = "OPEN"
	IncidentResolved IncidentStatus = "RESOLVED"
)

// Alert é o porte do tipo Alert (domain/types.ts).
type Alert struct {
	ID          string
	Service     string
	Description string
	Severity    Severity
	Status      AlertStatus
}

// Incident é o porte do tipo Incident (domain/types.ts). ResolvedAt/Summary
// usam ponteiro para representar a ausência de valor (equivalente a null em
// TS/Java) sem recorrer a um sentinel.
type Incident struct {
	ID         string
	Title      string
	Service    string
	Severity   Severity
	Status     IncidentStatus
	CreatedAt  int64
	ResolvedAt *int64
	Summary    *string
}

// Resolved retorna uma cópia do incidente marcada como resolvida, igual ao
// método Incident.resolved(...) do porte Java / à função pura equivalente em
// TS.
func (i Incident) Resolved(resolvedAt int64, summary string) Incident {
	r := i
	r.Status = IncidentResolved
	r.ResolvedAt = &resolvedAt
	r.Summary = &summary
	return r
}

// Runbook é o porte do tipo Runbook (domain/types.ts).
type Runbook struct {
	Service string
	Content string
}

// ConversationMessage é o porte do tipo ConversationMessage (domain/types.ts).
type ConversationMessage struct {
	Role    string
	Content string
}

// ExecutionMetrics é o porte do tipo ExecutionMetrics (domain/types.ts).
type ExecutionMetrics struct {
	LLMCalls  int
	LatencyMs int64
}

// TraceEventType é o porte do union type de TraceEvent["type"] (domain/types.ts).
type TraceEventType string

const (
	TraceThought     TraceEventType = "thought"
	TraceAction      TraceEventType = "action"
	TraceObservation TraceEventType = "observation"
	TracePlan        TraceEventType = "plan"
	TraceAnswer      TraceEventType = "answer"
	TraceHandoff     TraceEventType = "handoff"
)

// TraceEvent é o porte simplificado de TraceEvent (domain/types.ts) -- só os
// campos usados pelo núcleo portado. To fica vazio quando não aplicável
// (equivalente a undefined/null nos originais).
type TraceEvent struct {
	Type    TraceEventType
	Node    string
	Content string
	To      string
}

// Construtores de conveniência, equivalentes aos helpers estáticos de
// TraceEvent no porte Java.

func Handoff(node, to, content string) TraceEvent {
	return TraceEvent{Type: TraceHandoff, Node: node, Content: content, To: to}
}

func Answer(node, content string) TraceEvent {
	return TraceEvent{Type: TraceAnswer, Node: node, Content: content}
}

func Plan(node, content string) TraceEvent {
	return TraceEvent{Type: TracePlan, Node: node, Content: content}
}

func Observation(node, content string) TraceEvent {
	return TraceEvent{Type: TraceObservation, Node: node, Content: content}
}

func Action(node, content string) TraceEvent {
	return TraceEvent{Type: TraceAction, Node: node, Content: content}
}

func Thought(node, content string) TraceEvent {
	return TraceEvent{Type: TraceThought, Node: node, Content: content}
}

// StrategyRunInput é o porte de StrategyRunInput (domain/types.ts).
type StrategyRunInput struct {
	Message string
	History []ConversationMessage
}

// NewStrategyRunInput é o porte de StrategyRunInput.of(message) (porte Java).
func NewStrategyRunInput(message string) StrategyRunInput {
	return StrategyRunInput{Message: message}
}

// StrategyResult é o porte de StrategyResult (domain/types.ts).
type StrategyResult struct {
	Answer  string
	Trace   []TraceEvent
	Metrics ExecutionMetrics
}

// ReasoningStrategy é o porte da interface ReasoningStrategy (domain/types.ts).
// Diferente do original (TS/Java, que propagam falha via exceção), o método
// Run retorna um error explícito -- idiomático em Go e mais robusto que o
// porte Java, que deixa uma ModelUnavailableException não tratada estourar
// até o chamador.
type ReasoningStrategy interface {
	Name() string
	Run(input StrategyRunInput) (StrategyResult, error)
}
