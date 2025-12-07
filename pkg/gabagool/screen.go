package gabagool

// ScreenResult represents the outcome of drawing a screen
type ScreenResult[T any] struct {
	Value    T
	ExitCode ExitCode
}

// Success creates a successful result
func Success[T any](value T) ScreenResult[T] {
	return ScreenResult[T]{Value: value, ExitCode: ExitCodeSuccess}
}

// Back creates a back/cancel result with an optional value
func Back[T any](value ...T) ScreenResult[T] {
	var v T
	if len(value) > 0 {
		v = value[0]
	}
	return ScreenResult[T]{Value: v, ExitCode: ExitCodeBack}
}

// WithCode creates a result with a custom exit code
func WithCode[T any](value T, code ExitCode) ScreenResult[T] {
	return ScreenResult[T]{Value: value, ExitCode: code}
}

// Screen represents a type-safe UI screen
type Screen[I any, O any] interface {
	Draw(input I) (ScreenResult[O], error)
}

// ScreenFunc is a function adapter for the Screen interface
type ScreenFunc[I any, O any] func(input I) (ScreenResult[O], error)

func (f ScreenFunc[I, O]) Draw(input I) (ScreenResult[O], error) {
	return f(input)
}

// NoInput is used for screens that don't require input
type NoInput struct{}

// NoOutput is used for screens that don't produce meaningful output
type NoOutput struct{}
