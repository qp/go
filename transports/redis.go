package transports

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Redis is the Redis implementation of the
// Transport interface. It provides all functionality
// necessary to fulfill the Transport contract through
// a Redis transport layer.
type Redis struct {
	pool       *redis.Pool
	listeners  map[string][]MessageFunc
	once       sync.Once
	kill       chan struct{}
	processing int32
}

// NewRedis creates a Redis instance ready for use
func NewRedis(url string) Transport {
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
	return &Redis{pool: pool,
		listeners: map[string][]MessageFunc{},
		kill:      make(chan struct{})}
}

// ListenFor instructs Redis to deliver a message for the given topic
func (r *Redis) ListenFor(topic string, callback MessageFunc) error {
	if atomic.LoadInt32(&r.processing) == 1 {
		return ErrProcessingStarted
	}
	r.listeners[topic] = append(r.listeners[topic], callback)
	return nil
}

// Send sends a message out to Redis
func (r *Redis) Send(topic string, message []byte) error {
	conn := r.pool.Get()
	_, err := conn.Do("LPUSH", topic, message)
	conn.Close()
	return err
}

// Start begins processing messages to/from Redis
func (r *Redis) Start() error {
	r.processMessages()
	return nil
}

// Stop stops processing messages immediately
func (r *Redis) Stop() {
	close(r.kill)
	r.pool.Close()
}

func (r *Redis) processMessages() {
	r.once.Do(func() {
		atomic.StoreInt32(&r.processing, 1)
		for t, cbs := range r.listeners {
			go func(topic string, callbacks []MessageFunc) {
				var data []byte
				for {
					select {
					case <-r.kill:
						return
					default:
						conn := r.pool.Get()
						reply, err := redis.Values(conn.Do("BRPOP", topic, "0"))
						if err != nil {
							if netErr, ok := err.(net.Error); ok {
								if netErr.Timeout() {
									conn.Close()
									continue
								}
							}
							conn.Close()
							return
						}
						if _, err := redis.Scan(reply, &topic, &data); err != nil {
							conn.Close()
							return
						}
						for _, cb := range callbacks {
							go cb(&BinaryMessage{topic: topic, data: data})
						}
						conn.Close()
					}
				}
			}(t, cbs)
		}
	})
}
