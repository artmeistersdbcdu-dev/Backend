package cache

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const timeInterval = time.Minute * 60 * 24 * 7

type Cache struct {
	mu     sync.RWMutex
	Users  map[uuid.UUID]UserCache
	Events map[uuid.UUID]EventCache
	Arts   map[uuid.UUID]ArtCache
}

var (
	cache *Cache
	once  sync.Once
)

func LoadCache() *Cache {
	once.Do(func() {
		cache = &Cache{
			Users:  make(map[uuid.UUID]UserCache),
			Events: make(map[uuid.UUID]EventCache),
			Arts:   make(map[uuid.UUID]ArtCache),
		}
	})
	return cache
}

func cleanUp() {

}
func init() {
	go cleanUp()
}
