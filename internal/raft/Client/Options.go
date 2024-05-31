package raft

type Option func(*Client)

func WithServiceReg() Option {
	return func(c *Client) {
		c.ServiceRegistry = &ServiceRegistry{}
	}
}

func WithDataStore() Option {
	return func(c *Client) {
		c.Store, _ = c.BuildStore()
		//TODO : Handle store creation failure
		c.ServiceRegistry.client = c
	}
}
