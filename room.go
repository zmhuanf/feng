package feng

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type IRoom interface {
	// 获取房间ID
	GetID() string
	// 删除用户
	RemoveUser(user IUser) error
	// 获取用户
	GetUser(id string) (IUser, error)
	// 获取所有用户
	GetAllUsers() []IUser
	// 获取用户数量
	GetUserCount() int
	// 获取页码
	GetPage() int
}

type room struct {
	// 房间ID
	id string
	// 用户列表
	users map[string]*user
	// 用户锁
	lock sync.RWMutex
	// 服务
	server *server
	// 是否是系统房间
	isSys bool
	// 当前页码
	page int
	// 房主
	host *user
}

func newRoom(server *server, isSys bool) *room {
	r := &room{
		id:     uuid.New().String(),
		server: server,
		users:  make(map[string]*user),
		isSys:  isSys,
	}
	server.addRoom(r, isSys)
	return r
}

func (r *room) RemoveUser(iuser IUser) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	// 检查用户是否存在
	u, ok := r.users[iuser.GetID()]
	if !ok {
		return fmt.Errorf("user %s not found", iuser.GetID())
	}
	// 房主会直接解散房间
	if iuser.GetID() == r.host.GetID() {
		for _, user := range r.users {
			user.setRoom(nil)
		}
		r.server.removeRoom(r.id, r.isSys)
		return nil
	}
	// 不是房主，移除用户
	delete(r.users, iuser.GetID())
	u.setRoom(nil)
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

func (r *room) GetUserCount() int {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return len(r.users)
}

func (r *room) GetPage() int {
	return r.page
}

func (r *room) setPage(page int) {
	r.page = page
}

func (r *room) addUser(iuser *user) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	// 检查用户是否存在
	if _, ok := r.users[iuser.GetID()]; ok {
		return fmt.Errorf("user %s already exists", iuser.GetID())
	}
	// 添加用户到房间
	r.users[iuser.GetID()] = iuser
	return nil
}
