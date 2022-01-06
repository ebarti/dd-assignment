package common

// Restartable defines a component that can be started and stopped
type Restartable interface {
	Start() error
	Stop()
	IsStopped() bool
}
