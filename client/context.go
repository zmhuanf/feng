package client

type IContext interface {
	GetClient() IClient
}

type fContext struct {
	client IClient
}

func (c *fContext) GetClient() IClient {
	return c.client
}

func newContext(client IClient) IContext {
	return &fContext{
		client: client,
	}
}
