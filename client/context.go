package client

type IContext interface {
	GetClient() IClient
}

type context struct {
	client IClient
}

func (c *context) GetClient() IClient {
	return c.client
}

func newContext(client IClient) IContext {
	return &context{
		client: client,
	}
}
