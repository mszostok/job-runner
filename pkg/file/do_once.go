package file

import "sync"

type doOnlyOnce struct {
	err    error
	closer func() error
	once   sync.Once
}

func newDoOnlyOnce(closer func() error) *doOnlyOnce {
	return &doOnlyOnce{
		closer: closer,
		once:   sync.Once{},
	}
}

func (c *doOnlyOnce) Close() {
	c.once.Do(func() {
		c.err = c.closer()
	})
}

func (c *doOnlyOnce) Err() error {
	return c.err
}
