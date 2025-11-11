package feng

import (
	"fmt"
	"sync"
)

type IRoom interface {
	// 获取房间ID
	GetID() string
	// 添加用户
	AddUser(user IUser) error
	// 删除用户
	RemoveUser(user IUser) error
	// 获取用户
	GetUser(id string) (IUser, error)
	// 获取所有用户
	GetAllUsers() []IUser
}

type room struct {
	// 房间ID
	id string
	// 用户列表
	users map[string]IUser
	// 房间锁
	lock sync.RWMutex
}

func NewRoom(id string) IRoom {
	return &room{
		id:    id,
		users: make(map[string]IUser),
	}
}

func (r *room) AddUser(user IUser) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.users[user.GetID()] = user
	user.SetRoom(r)
	return nil
}

func (r *room) RemoveUser(user IUser) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.users, user.GetID())
	user.SetRoom(nil)
	return nil
}

func (r *room) GetUser(id string) (IUser, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return user, nil
}

func (r *room) GetAllUsers() []IUser {
	r.lock.RLock()
	defer r.lock.RUnlock()
	users := make([]IUser, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users
}

func (r *room) GetID() string {
	return r.id
}
