package bus

// Logger is an interface that allows passing of custom logging
type Logger interface {
	Info(msg string, args ...any)
}
