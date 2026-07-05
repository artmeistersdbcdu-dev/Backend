package cache

import "github.com/google/uuid"

type Cache struct {
	Users  map[uuid.UUID]UserCache
	Events map[uuid.UUID]EventCache
	Arts   map[uuid.UUID]ArtCache
}
