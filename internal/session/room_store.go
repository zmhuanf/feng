package session

import (
	"errors"
	"sync"

	"github.com/zmhuanf/feng/internal/core"
)

type RoomStoreImpl struct {
	pageSize int
	rooms    map[string]*Room
	index    map[int]map[string]*Room
	lock     sync.RWMutex
}

func NewRoomStore(pageSize int) *RoomStoreImpl {
	return &RoomStoreImpl{
		pageSize: pageSize,
		rooms:    make(map[string]*Room),
		index:    make(map[int]map[string]*Room),
	}
}

func (s *RoomStoreImpl) CreateRoom() *Room {
	room := NewRoom(s)
	_ = s.AddRoom(room)
	return room
}

func (s *RoomStoreImpl) AddRoom(room *Room) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.rooms[room.ID()]; ok {
		return errors.New("room already exists")
	}
	s.rooms[room.ID()] = room
	page := s.nextPageLocked()
	if s.index[page] == nil {
		s.index[page] = make(map[string]*Room)
	}
	s.index[page][room.ID()] = room
	room.SetPage(page)
	return nil
}

func (s *RoomStoreImpl) Room(id string) (core.Room, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	room, ok := s.rooms[id]
	if !ok {
		return nil, errors.New("room not found")
	}
	return room, nil
}

func (s *RoomStoreImpl) Rooms() []core.Room {
	s.lock.RLock()
	defer s.lock.RUnlock()
	rooms := make([]core.Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

func (s *RoomStoreImpl) RoomsByPage(page int) []core.Room {
	s.lock.RLock()
	defer s.lock.RUnlock()
	rooms := make([]core.Room, 0, len(s.index[page]))
	for _, room := range s.index[page] {
		rooms = append(rooms, room)
	}
	return rooms
}

func (s *RoomStoreImpl) RemoveRoom(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	room, ok := s.rooms[id]
	if !ok {
		return errors.New("room not found")
	}
	delete(s.rooms, id)
	delete(s.index[room.Page()], id)
	return nil
}

func (s *RoomStoreImpl) nextPageLocked() int {
	for page := 0; ; page++ {
		if len(s.index[page]) < s.pageSize {
			return page
		}
	}
}
