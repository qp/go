package request

import (
	"net"
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
	timeout   time.Duration
}

// MakeRedis initializes a new Redis transport instance
func MakeRedis(url string) transports.RequestTransport {
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
		timeout:   5 * time.Second,
	}
}

// Send sends a message to Redis and returns an error
// if there was a problem sending.
func (r *Redis) Send(to string, data []byte) error {
	if atomic.LoadUint32(&r.running) == 0 {
		return transports.ErrTransportStopped
	}
	conn := r.pool.Get()
	_, err := conn.Do("LPUSH", to, data)
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

// SetTimeout sets the timeout to the given value.
// This timeout is used when gracefully shutting down the
// transport. In-flight requests will have this much time
// to complete before being abandoned.
// The default timeout value is 5 seconds.
func (r *Redis) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

func (r *Redis) processMessages() {
	r.lock.Lock()
	listeners := r.listeners
	r.lock.Unlock()
	for _, c := range listeners {
		go func(channel string) {
			var data []byte
			done := make(chan struct{})
			for {
				conn := r.pool.Get()
				select {
				case <-r.kill:
					go func() {
						select {
						case <-time.After(r.timeout):
							done <- struct{}{}
						}
					}()
				case <-done:
					conn.Close()
					return
				default:
					// BRPOP on the channel to wait for a new message
					message, err := redis.Values(conn.Do("BRPOP", channel, "1"))
					if err != nil {
						// Did we get a timeout? That's fine. Continue.
						if netErr, ok := err.(net.Error); ok {
							if netErr.Timeout() {
								continue
							}
						}
						// Not a timeout? Something went wrong.
						// TODO: Log this out.. maybe fire a metric to the logging endpoint
						conn.Close()
						return
					}
					if _, err := redis.Scan(message, &channel, &data); err != nil {
						// there was an error decoding the message into the data field
						// TODO: Log this out.. maybe fire a metric to the logging endpoint
						conn.Close()
						return
					}
					go r.callback(&transports.BinaryMessage{Channel: channel, Data: data})
					conn.Close()
				}
			}
		}(c)
	}
}
