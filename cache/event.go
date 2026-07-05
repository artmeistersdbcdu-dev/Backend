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
func (c *Cache) GetEvent(id uuid.UUID) *Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	event, ok := c.Events[id]
	if !ok {
		return nil
	}
	event.ExpireAt = time.Now().Add(timeInterval)
	c.Events[id] = event
	return &event.Event
}
func (c *Cache) DeleteEvent(id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Events, id)
}
func (c *Cache) EventCleanUp(curr time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, item := range c.Events {
		if !curr.Before(item.ExpireAt) {
			delete(c.Events, id)
		}
	}
}
