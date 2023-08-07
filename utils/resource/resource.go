package resource

import (
	"io"
	"sync"
)

type closeableResource struct {
	io.Closer
	mu                sync.RWMutex
	closeableResource io.Closer
	closed            bool
	description       string
}

func (c *closeableResource) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

func (c *closeableResource) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closeableResource != nil {
		err := c.closeableResource.Close()
		if err != nil {
			return err
		}
	}
	c.closed = true
	c.closeableResource = nil
	return nil
}

func (c *closeableResource) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.description
}

// NewCloseableResource returns a Closeable resource.
func NewCloseableResource(resource io.Closer, description string) ICloseableResource {
	return &closeableResource{
		closeableResource: resource,
		closed:            false,
		description:       description,
	}
}

type closeableNilResource struct {
}

func (c *closeableNilResource) Close() error {
	return nil
}

func (c *closeableNilResource) String() string {
	return "non closeable resource"
}

func (c *closeableNilResource) IsClosed() bool {
	return false
}

// NewNonCloseableResource returns a resource which cannot be closed.
func NewNonCloseableResource() ICloseableResource {
	return &closeableNilResource{}
}
