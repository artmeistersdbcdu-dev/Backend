package cache

import (
	"time"

	"github.com/google/uuid"
)

func (c *Cache) GetUser(id uuid.UUID) *User {
	c.mu.Lock()
	defer c.mu.Unlock()
	user, ok := c.Users[id]
	if !ok {
		return nil
	}
	user.ExpireAt = time.Now().Add(timeInterval)
	c.Users[id] = user
	return &user.User
}
func (c *Cache) DeleteUser(id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Users, id)
}
func (c *Cache) SetUser(id uuid.UUID, user User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Users[id] = UserCache{
		User:     user,
		ExpireAt: time.Now().Add(timeInterval),
	}

}
func (c *Cache) UserCleanUp(curr time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, item := range c.Users {
		if !curr.Before(item.ExpireAt) {
			delete(c.Users, id)
		}
	}
}
