package natsx

import (
	"time"

	nats "github.com/nats-io/nats.go"
)

// Client is a tiny facade for the subset of nats.Conn we use.
type Client interface {
	Publish(subj string, data []byte) error
	Request(subj string, data []byte, timeout time.Duration) (*nats.Msg, error)
	Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error)
}

// Conn wraps *nats.Conn to satisfy Client.
type Conn struct{ C *nats.Conn }

func New(c *nats.Conn) *Conn { return &Conn{C: c} }

func (c *Conn) Publish(subj string, data []byte) error { return c.C.Publish(subj, data) }
func (c *Conn) Request(subj string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	return c.C.Request(subj, data, timeout)
}

func (c *Conn) Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error) {
	return c.C.Subscribe(subj, cb)
}
