package flowcontroller

type Work func() error

type FlowController interface {
	Do(Work) error
}
