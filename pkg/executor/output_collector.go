package executor

import "io"

type collector struct {
	dataCopy   []byte
	nextWriter io.Writer
}

func (c *collector) Write(p []byte) (n int, err error) {
	c.dataCopy = append(c.dataCopy, p...)
	return c.nextWriter.Write(p)
}

func (c *collector) Data() []byte {
	return c.dataCopy
}

func newCollector(nextWriter io.Writer) *collector {
	return &collector{
		dataCopy:   []byte{},
		nextWriter: nextWriter,
	}
}
