package session

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/zmhuanf/feng/internal/core"
)

type RoomStore interface {
	RemoveRoom(id string) error
}

type Room struct {
	id    string
	users map[string]*User
	lock  sync.RWMutex
	store RoomStore
	page  int
	host  *User
}

func NewRoom(store RoomStore) *Room {
	return &Room{
		id:    uuid.New().String(),
		users: make(map[string]*User),
		store: store,
	}
}

func (r *Room) ID() string { return r.id }

func (r *Room) RemoveUser(user core.User) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	u, ok := user.(*User)
	if !ok {
		return fmt.Errorf("invalid user type")
	}
	if _, ok := r.users[u.ID()]; !ok {
		return fmt.Errorf("user %s not found", u.ID())
	}
	if r.host != nil && r.host.ID() == u.ID() {
		for _, item := range r.users {
			item.setRoom(nil)
		}
		r.users = make(map[string]*User)
		return r.store.RemoveRoom(r.id)
	}
	delete(r.users, u.ID())
	u.setRoom(nil)
	return nil
}

func (r *Room) User(id string) (core.User, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return u, nil
}

func (r *Room) Users() []core.User {
	r.lock.RLock()
	defer r.lock.RUnlock()
	users := make([]core.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users
}

func (r *Room) UserCount() int {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return len(r.users)
}

func (r *Room) Page() int { return r.page }

func (r *Room) SetPage(page int) { r.page = page }

func (r *Room) AddUser(user *User) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, ok := r.users[user.ID()]; ok {
		return fmt.Errorf("user %s already exists", user.ID())
	}
	if r.host == nil {
		r.host = user
	}
	r.users[user.ID()] = user
	return nil
}
