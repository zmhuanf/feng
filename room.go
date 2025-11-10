package feng

type IRoom interface {
	// 广播
	Broadcast(msg any) error
	// 除了自己以外的广播
	BroadcastExceptSelf(msg any) error
}

type room struct {
	// 房间ID
	id string
	// 用户列表
	users map[string]IUser
}

func (r *room) Broadcast(msg any) error {
	return nil
}

func (r *room) BroadcastExceptSelf(msg any) error {
	return nil
}
