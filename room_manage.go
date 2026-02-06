package feng

import (
	"errors"
	"sync"
)

type roomManage struct {
	config    *serverConfig
	rooms     map[string]*room
	roomsLock sync.RWMutex
	index     map[int]map[string]*room
	indexLock sync.RWMutex
}

func newRoomManage(config *serverConfig) *roomManage {
	return &roomManage{
		config: config,
		rooms:  make(map[string]*room),
		index:  make(map[int]map[string]*room),
	}
}

// 获取房间
func (r *roomManage) getRoom(id string) (IRoom, error) {
	r.roomsLock.RLock()
	defer r.roomsLock.RUnlock()
	room, ok := r.rooms[id]
	if !ok {
		return nil, errors.New("room not found")
	}
	return room, nil
}

// 获取所有房间
func (r *roomManage) getAllRooms() []IRoom {
	r.roomsLock.RLock()
	defer r.roomsLock.RUnlock()
	rooms := make([]IRoom, 0, len(r.rooms))
	for _, room := range r.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// 获取指定页房间
func (r *roomManage) getRoomsByPage(page int) []IRoom {
	r.indexLock.RLock()
	defer r.indexLock.RUnlock()
	rooms := make([]IRoom, 0)
	pageRooms, ok := r.index[page]
	if !ok {
		return rooms
	}
	for _, room := range pageRooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// 添加房间
func (r *roomManage) addRoom(iroom *room) error {
	r.roomsLock.Lock()
	defer r.roomsLock.Unlock()
	r.indexLock.Lock()
	defer r.indexLock.Unlock()

	// 检查房间是否已存在
	if _, ok := r.rooms[iroom.GetID()]; ok {
		return errors.New("room already exists")
	}
	// 添加房间到房间列表
	r.rooms[iroom.GetID()] = iroom
	// 添加房间到索引
	page := 0
	for {
		if _, ok := r.index[page]; !ok {
			r.index[page] = make(map[string]*room)
		}
		if len(r.index[page]) < r.config.PageSize {
			r.index[page][iroom.GetID()] = iroom
			break
		}
		page++
	}
	// 设置房间页码
	iroom.setPage(page)
	return nil
}

// 移除房间
func (r *roomManage) removeRoom(id string) error {
	r.roomsLock.Lock()
	defer r.roomsLock.Unlock()
	r.indexLock.Lock()
	defer r.indexLock.Unlock()
	// 检查房间是否存在
	iroom, ok := r.rooms[id]
	if !ok {
		return errors.New("room not found")
	}
	// 从房间列表中移除房间
	delete(r.rooms, id)
	// 从索引中移除房间
	index, ok := r.index[iroom.GetPage()]
	if !ok {
		r.index[iroom.GetPage()] = make(map[string]*room)
	}
	delete(index, id)
	return nil
}
