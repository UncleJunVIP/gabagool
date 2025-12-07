package gabagool

// FSMBuilder provides a fluent API for building FSMs
type FSMBuilder struct {
	fsm *FSM
}

// NewFSMBuilder creates a new builder
func NewFSMBuilder() *FSMBuilder {
	return &FSMBuilder{fsm: NewFSM()}
}

// StartWith sets the initial node
func (b *FSMBuilder) StartWith(name string) *FSMBuilder {
	b.fsm.SetInitial(name)
	return b
}

// WithData adds initial context data
func (b *FSMBuilder) WithData(key string, value any) *FSMBuilder {
	b.fsm.ctx.Set(key, value)
	return b
}

// On starts a transition definition
func (b *FSMBuilder) On(from string, code ExitCode) *TransitionBuilder {
	return &TransitionBuilder{builder: b, from: from, code: code}
}

// Build returns the configured FSM
func (b *FSMBuilder) Build() *FSM {
	return b.fsm
}

// TransitionBuilder for fluent transition creation
type TransitionBuilder struct {
	builder *FSMBuilder
	from    string
	code    ExitCode
	hook    func(ctx *FSMContext) error
}

// Before adds a hook before the transition
func (t *TransitionBuilder) Before(hook func(ctx *FSMContext) error) *TransitionBuilder {
	t.hook = hook
	return t
}

// GoTo sets the destination node
func (t *TransitionBuilder) GoTo(to string) *FSMBuilder {
	t.builder.fsm.transitions = append(t.builder.fsm.transitions, Transition{
		FromNode:         t.from,
		ExitCode:         t.code,
		ToNode:           to,
		BeforeTransition: t.hook,
	})
	return t.builder
}

// Exit marks this as an exit transition
func (t *TransitionBuilder) Exit() *FSMBuilder {
	return t.GoTo("")
}

// RegisterScreenWithHandler registers a screen with input factory and output handler
func RegisterScreenWithHandler[I any, O any](
	b *FSMBuilder,
	name string,
	screen Screen[I, O],
	inputFactory func(ctx *FSMContext) I,
	outputHandler func(ctx *FSMContext, output O),
) *FSMBuilder {
	AddNodeWithHandler(b.fsm, name, screen, inputFactory, outputHandler)
	return b
}
