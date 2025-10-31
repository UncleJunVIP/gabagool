package gabagool

var globalInputProcessor *InputProcessor

func InitInputProcessor() {
	globalInputProcessor = NewInputProcessor()
}

func GetInputProcessor() *InputProcessor {
	if globalInputProcessor == nil {
		InitInputProcessor()
	}
	return globalInputProcessor
}

func SetInputMapping(mapping *InputMapping) {
	GetInputProcessor().SetMapping(mapping)
}
