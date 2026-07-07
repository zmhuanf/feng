package session

import (
	"errors"
	"sync"

	"github.com/zmhuanf/feng/internal/core"
)

type UserStore struct {
	pageSize int
	users    map[string]*User
	index    map[int]map[string]*User
	lock     sync.RWMutex
}

func NewUserStore(pageSize int) *UserStore {
	return &UserStore{
		pageSize: pageSize,
		users:    make(map[string]*User),
		index:    make(map[int]map[string]*User),
	}
}

func (s *UserStore) Add(user *User) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.users[user.ID()]; ok {
		return errors.New("user already exists")
	}
	s.users[user.ID()] = user
	page := s.nextPageLocked()
	if s.index[page] == nil {
		s.index[page] = make(map[string]*User)
	}
	s.index[page][user.ID()] = user
	user.setPage(page)
	return nil
}

func (s *UserStore) User(id string) (core.User, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	user, ok := s.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserStore) Users() []core.User {
	s.lock.RLock()
	defer s.lock.RUnlock()
	users := make([]core.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

func (s *UserStore) UsersByPage(page int) []core.User {
	s.lock.RLock()
	defer s.lock.RUnlock()
	users := make([]core.User, 0, len(s.index[page]))
	for _, user := range s.index[page] {
		users = append(users, user)
	}
	return users
}

func (s *UserStore) Remove(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	user, ok := s.users[id]
	if !ok {
		return errors.New("user not found")
	}
	delete(s.users, id)
	delete(s.index[user.Page()], id)
	return nil
}

func (s *UserStore) nextPageLocked() int {
	for page := 0; ; page++ {
		if len(s.index[page]) < s.pageSize {
			return page
		}
	}
}
