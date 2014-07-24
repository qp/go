package redis

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/qp/go"
)

// events implements the Transport interface using
// r as the underlying transport technology.
type events struct {
	pool      *redis.Pool
	callback  qp.MessageFunc
	listeners []string
	lock      sync.Mutex
	once      sync.Once
	kill      chan struct{}
	running   uint32
}

// ensure events is a valid qp.EventTransport
var _ qp.PubSubTransport = (*events)(nil)

// NewPubSubTransport initializes a Redis qp.EventTransport.
func NewPubSubTransport(url string) qp.PubSubTransport {
	var pool = &redis.Pool{
		MaxIdle:     8,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", url, 1*time.Second, 0, 0)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
	return &events{
		pool:      pool,
		listeners: []string{},
		kill:      make(chan struct{}),
	}
}

// Send publishes a message to r and returns an error
// if there was a problem sending.
func (r *events) Send(to string, data []byte) error {
	if atomic.LoadUint32(&r.running) == 0 {
		return qp.ErrTransportStopped
	}
	conn := r.pool.Get()
	_, err := conn.Do("PUBLISH", to, data)
	conn.Close()
	return err
}

// ListenFor instructs the transport to register for
// messages on the given channel.
func (r *events) ListenFor(channel string) {
	r.lock.Lock()
	r.listeners = append(r.listeners, channel)
	r.lock.Unlock()
}

// ListenForChildren instructs the transport to register for
// messages on the given channel and its children
func (r *events) ListenForChildren(channel string) {
	channel += "*"
	r.lock.Lock()
	r.listeners = append(r.listeners, channel)
	r.lock.Unlock()
}

// OnMessage sets the callback function to be called whenever
// a message is received.
func (r *events) OnMessage(messageFunc qp.MessageFunc) {
	r.callback = messageFunc
}

// Start spins up the r transport processing system
// and begins processing messages.
func (r *events) Start() error {
	r.once.Do(func() {
		go r.processMessages()
		atomic.StoreUint32(&r.running, 1)
	})
	return nil
}

// Stop spins down the r transport processing system
func (r *events) Stop() {
	close(r.kill)
	r.kill = make(chan struct{})
	r.once = sync.Once{}
	atomic.StoreUint32(&r.running, 0)
}

// SetTimeout is a no-op for event transports as there
// is no in-flight request to wait for.
func (r *events) SetTimeout(timeout time.Duration) {
}

func (r *events) processMessages() {
	r.lock.Lock()
	listeners := r.listeners
	r.lock.Unlock()
	for _, c := range listeners {
		go func(channel string) {
			conn := r.pool.Get()
			psc := redis.PubSubConn{Conn: conn}
			psc.PSubscribe(channel)
		loop:
			for {
				select {
				case <-r.kill:
					break loop
				default:
					switch v := psc.Receive().(type) {
					case redis.PMessage:
						go r.callback(&qp.Message{Source: v.Channel, Data: v.Data})
					case error:
						break loop
					}
				}
			}
			psc.PUnsubscribe(channel)
			conn.Close()
		}(c)
	}
}
