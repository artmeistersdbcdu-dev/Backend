package cache

import (
	"time"

	"github.com/google/uuid"
)

func (c *Cache) SetEvent(id uuid.UUID, event Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Events[id] = EventCache{
		Event:    event,
		ExpireAt: time.Now().Add(timeInterval),
	}
}
func (c *Cache) GetEvent(id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
}
func (c *Cache) DeleteEvent(id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Events, id)
}
