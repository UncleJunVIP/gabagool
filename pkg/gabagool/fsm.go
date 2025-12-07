package gabagool

import (
	"fmt"
)

// ExitCode represents the result of a screen draw operation
type ExitCode int

const (
	ExitCodeSuccess  ExitCode = 0
	ExitCodeBack     ExitCode = 1
	ExitCodeCancel   ExitCode = 2
	ExitCodeError    ExitCode = -1
	ExitCodeSettings ExitCode = 4
	ExitCodeSearch   ExitCode = 5
	// Custom exit codes can start from 100
)

// FSMContext holds shared state between screens during FSM execution
type FSMContext struct {
	data map[string]any
}

// NewFSMContext creates a new context
func NewFSMContext() *FSMContext {
	return &FSMContext{data: make(map[string]any)}
}

// Set stores a value in the context
func (c *FSMContext) Set(key string, value any) {
	c.data[key] = value
}

// Get retrieves a typed value from the context
func Get[T any](c *FSMContext, key string) (T, bool) {
	val, ok := c.data[key]
	if !ok {
		var zero T
		return zero, false
	}
	typed, ok := val.(T)
	return typed, ok
}

// MustGet retrieves a typed value or panics
func MustGet[T any](c *FSMContext, key string) T {
	val, ok := Get[T](c, key)
	if !ok {
		panic(fmt.Sprintf("key %q not found or wrong type", key))
	}
	return val
}

// Node represents a node in the FSM that can be executed
type Node interface {
	execute(ctx *FSMContext) (exitCode ExitCode, err error)
	name() string
}

// ScreenNode wraps a typed screen in the FSM
type ScreenNode[I any, O any] struct {
	nodeName      string
	screen        Screen[I, O]
	inputFactory  func(ctx *FSMContext) I
	outputHandler func(ctx *FSMContext, output O) // Called with the output after Draw
}

func (n *ScreenNode[I, O]) name() string {
	return n.nodeName
}

func (n *ScreenNode[I, O]) execute(ctx *FSMContext) (ExitCode, error) {
	var input I
	if n.inputFactory != nil {
		input = n.inputFactory(ctx)
	}

	result, err := n.screen.Draw(input)
	if err != nil {
		return ExitCodeError, err
	}

	if n.outputHandler != nil {
		n.outputHandler(ctx, result.Value)
	}

	return result.ExitCode, nil
}

// Transition defines a state transition
type Transition struct {
	FromNode         string
	ExitCode         ExitCode
	ToNode           string
	BeforeTransition func(ctx *FSMContext) error
}

// FSM represents the finite state machine
type FSM struct {
	nodes       map[string]Node
	transitions []Transition
	initialNode string
	ctx         *FSMContext
}

// NewFSM creates a new FSM
func NewFSM() *FSM {
	return &FSM{
		nodes:       make(map[string]Node),
		transitions: []Transition{},
		ctx:         NewFSMContext(),
	}
}

// AddNodeWithHandler adds a screen node with input factory and output handler
func AddNodeWithHandler[I any, O any](
	fsm *FSM,
	name string,
	screen Screen[I, O],
	inputFactory func(ctx *FSMContext) I,
	outputHandler func(ctx *FSMContext, output O),
) {
	fsm.nodes[name] = &ScreenNode[I, O]{
		nodeName:      name,
		screen:        screen,
		inputFactory:  inputFactory,
		outputHandler: outputHandler,
	}
}

// SetInitial sets the starting node
func (f *FSM) SetInitial(name string) *FSM {
	f.initialNode = name
	return f
}

// Context returns the FSM context
func (f *FSM) Context() *FSMContext {
	return f.ctx
}

// Run executes the FSM
func (f *FSM) Run() error {
	currentNode := f.initialNode

	for {
		node, exists := f.nodes[currentNode]
		if !exists {
			return fmt.Errorf("node not found: %s", currentNode)
		}

		exitCode, err := node.execute(f.ctx)
		if err != nil {
			return fmt.Errorf("error in node %s: %w", currentNode, err)
		}

		var matched *Transition
		for i := range f.transitions {
			t := &f.transitions[i]
			if t.FromNode == currentNode && t.ExitCode == exitCode {
				matched = t
				break
			}
		}

		if matched == nil {
			return nil
		}

		if matched.BeforeTransition != nil {
			if err := matched.BeforeTransition(f.ctx); err != nil {
				return fmt.Errorf("transition hook error: %w", err)
			}
		}

		if matched.ToNode == "" {
			return nil
		}

		currentNode = matched.ToNode
	}
}
