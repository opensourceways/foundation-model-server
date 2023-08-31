package flowcontrollerimpl

import (
	"github.com/opensourceways/foundation-model-server/allerror"
	"github.com/opensourceways/foundation-model-server/inferenceqa/domain/flowcontroller"
)

func Init(n int) *flowControllerImpl {
	c := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		c <- struct{}{}
	}

	return &flowControllerImpl{c}
}

type flowControllerImpl struct {
	c chan struct{}
}

func (s *flowControllerImpl) Do(w flowcontroller.Work) error {
	select {
	case e := <-s.c:
		err := w()
		s.c <- e

		return err

	default:
		return allerror.New(allerror.ErrorCodeTooManyRequest, "")
	}
}
