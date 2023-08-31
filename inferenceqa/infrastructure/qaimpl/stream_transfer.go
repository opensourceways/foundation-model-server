package qaimpl

import (
	"errors"
	"io"
)

type streamTransfer struct {
	input io.Reader
}

func (impl *streamTransfer) readAndWriteOnce(output io.Writer) (bool, error) {
	buf := make([]byte, 1<<10)

	n, err := impl.input.Read(buf)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return false, err
		}

		return true, nil
	}

	if n > 0 {
		err = impl.write(output, buf[:n])
	}

	return false, err
}

func (impl *streamTransfer) write(output io.Writer, data []byte) error {
	for v := data; len(v) > 0; {
		n, err := output.Write(v)
		if err == nil {
			return nil
		}

		if n == 0 {
			return err
		}

		v = v[n:]
	}

	return nil
}
