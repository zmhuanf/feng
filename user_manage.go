package feng

import (
	"errors"
	"sync"
)

type userManage struct {
	config    *serverConfig
	users     map[string]*user
	usersLock sync.RWMutex
	index     map[int]map[string]*user
	indexLock sync.RWMutex
}

func newUserManage(config *serverConfig) *userManage {
	return &userManage{
		config: config,
		users:  make(map[string]*user),
		index:  make(map[int]map[string]*user),
	}
}

// 获取用户
func (u *userManage) getUser(id string) (IUser, error) {
	u.usersLock.RLock()
	defer u.usersLock.RUnlock()
	user, ok := u.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// 获取所有用户
func (u *userManage) getAllUsers() []IUser {
	u.usersLock.RLock()
	defer u.usersLock.RUnlock()
	users := make([]IUser, 0, len(u.users))
	for _, user := range u.users {
		users = append(users, user)
	}
	return users
}

// 获取指定页用户
func (u *userManage) getUsersByPage(page int) []IUser {
	u.indexLock.RLock()
	defer u.indexLock.RUnlock()
	users := make([]IUser, 0)
	pageUsers, ok := u.index[page]
	if !ok {
		return users
	}
	for _, user := range pageUsers {
		users = append(users, user)
	}
	return users
}

// 添加用户
func (u *userManage) addUser(iuser *user) error {
	u.usersLock.Lock()
	defer u.usersLock.Unlock()
	u.indexLock.Lock()
	defer u.indexLock.Unlock()

	// 检查用户是否已存在
	if _, ok := u.users[iuser.GetID()]; ok {
		return errors.New("user already exists")
	}
	// 添加用户到用户列表
	u.users[iuser.GetID()] = iuser
	// 添加用户到索引
	page := 0
	for {
		if _, ok := u.index[page]; !ok {
			u.index[page] = make(map[string]*user)
		}
		if len(u.index[page]) < u.config.PageSize {
			u.index[page][iuser.GetID()] = iuser
			break
		}
		page++
	}
	// 设置用户页码
	iuser.setPage(page)
	return nil
}

// 移除用户
func (u *userManage) removeUser(id string) error {
	u.usersLock.Lock()
	defer u.usersLock.Unlock()
	u.indexLock.Lock()
	defer u.indexLock.Unlock()
	// 检查用户是否存在
	iuser, ok := u.users[id]
	if !ok {
		return errors.New("user not found")
	}
	// 从用户列表中移除用户
	delete(u.users, id)
	// 从索引中移除用户
	index, ok := u.index[iuser.GetPage()]
	if !ok {
		u.index[iuser.GetPage()] = make(map[string]*user)
	}
	delete(index, id)
	return nil
}
