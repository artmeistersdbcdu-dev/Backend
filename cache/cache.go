package cache

import (
	"sync"

	"github.com/google/uuid"
)

type Cache struct {
	mu     sync.RWMutex
	Users  map[uuid.UUID]UserCache
	Events map[uuid.UUID]EventCache
	Arts   map[uuid.UUID]ArtCache
}

func GetUser(id uuid.UUID)            {}
func DeleteUser(id uuid.UUID)         {}
func SetUser(id uuid.UUID, user User) {}

func GetArt(id uuid.UUID)          {}
func DeleteArt(id uuid.UUID)       {}
func SetArt(id uuid.UUID, art Art) {}

func SetEvent(id uuid.UUID, event Event) {}
func GetEvent(id uuid.UUID)              {}
func DeleteEvent(id uuid.UUID)           {}
