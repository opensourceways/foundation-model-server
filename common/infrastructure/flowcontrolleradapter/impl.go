package flowcontrolleradapter

import (
	"github.com/opensourceways/foundation-model-server/allerror"
	"github.com/opensourceways/foundation-model-server/common/domain/flowcontroller"
)

func Init(n int) *flowControllerAdapter {
	c := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		c <- struct{}{}
	}

	return &flowControllerAdapter{c}
}

type flowControllerAdapter struct {
	c chan struct{}
}

func (s *flowControllerAdapter) Do(w flowcontroller.Work) error {
	select {
	case e := <-s.c:
		err := w()
		s.c <- e

		return err

	default:
		return allerror.New(allerror.ErrorCodeTooManyRequest, "")
	}
}
