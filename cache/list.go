package cache

import "time"

func (c *Cache) GetList(key string) any {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	list, ok := c.Lists[key]
	if !ok {
		return nil
	}
	list.ExpireAt = time.Now().Add(timeInterval)
	c.Lists[key] = list
	return list.Data
}

func (c *Cache) SetList(key string, data any) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Lists[key] = ListCache{
		Data:     data,
		ExpireAt: time.Now().Add(timeInterval),
	}
}

func (c *Cache) DeleteList(key string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Lists, key)
}

func (c *Cache) ListCleanUp(curr time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.Lists {
		if !curr.Before(item.ExpireAt) {
			delete(c.Lists, key)
		}
	}
}
