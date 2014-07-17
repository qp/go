package event

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/qp/go/transports"
)

// Redis implements the Transport interface using
// Redis as the underlying transport technology.
type Redis struct {
	pool      *redis.Pool
	callback  transports.MessageFunc
	listeners []string
	lock      sync.Mutex
	once      sync.Once
	kill      chan struct{}
	running   uint32
}

// MakeRedis initializes a new Redis transport instance
func MakeRedis(url string) transports.EventTransport {
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
	return &Redis{
		pool:      pool,
		listeners: []string{},
		kill:      make(chan struct{}),
	}
}

// Send publishes a message to Redis and returns an error
// if there was a problem sending.
func (r *Redis) Send(to string, data []byte) error {
	if atomic.LoadUint32(&r.running) == 0 {
		return transports.ErrTransportStopped
	}
	conn := r.pool.Get()
	_, err := conn.Do("PUBLISH", to, data)
	conn.Close()
	return err
}

// ListenFor instructs the transport to register for
// messages on the given channel.
func (r *Redis) ListenFor(channel string) {
	r.lock.Lock()
	r.listeners = append(r.listeners, channel)
	r.lock.Unlock()
}

// ListenForChildren instructs the transport to register for
// messages on the given channel and its children
func (r *Redis) ListenForChildren(channel string) {
	channel += "*"
	r.lock.Lock()
	r.listeners = append(r.listeners, channel)
	r.lock.Unlock()
}

// OnMessage sets the callback function to be called whenever
// a message is received.
func (r *Redis) OnMessage(messageFunc transports.MessageFunc) {
	r.callback = messageFunc
}

// Start spins up the Redis transport processing system
// and begins processing messages.
func (r *Redis) Start() error {
	r.once.Do(func() {
		go r.processMessages()
		atomic.StoreUint32(&r.running, 1)
	})
	return nil
}

// Stop spins down the Redis transport processing system
func (r *Redis) Stop() {
	close(r.kill)
	r.kill = make(chan struct{})
	r.once = sync.Once{}
	atomic.StoreUint32(&r.running, 0)
}

// SetTimeout is a no-op for event transports as there
// is no in-flight request to wait for.
func (r *Redis) SetTimeout(timeout time.Duration) {
}

func (r *Redis) processMessages() {
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
						go r.callback(&transports.BinaryMessage{Channel: v.Channel, Data: v.Data})
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
