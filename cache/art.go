package cache

import (
	"time"

	"github.com/google/uuid"
)

func (c *Cache) GetArt(id uuid.UUID) *Art {
	c.mu.Lock()
	defer c.mu.Unlock()

	art, ok := c.Arts[id]
	if !ok {
		return nil
	}
	art.ExpireAt = time.Now().Add(timeInterval)
	c.Arts[id] = art
	return &art.Art
}
func (c *Cache) DeleteArt(id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Arts, id)
}
func (c *Cache) SetArt(id uuid.UUID, art Art) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Arts[id] = ArtCache{
		Art:      art,
		ExpireAt: time.Now().Add(timeInterval),
	}

}
