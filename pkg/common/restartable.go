package common

// Restartable defines a service
type Restartable interface {
	Start() error
	Stop()
}
