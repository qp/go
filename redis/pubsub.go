package redis

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/qp/go"
	"github.com/stretchr/pat/stop"
)

// PubSub represents a qp.PubSubTransport.
type PubSub struct {
	pool     *redis.Pool
	handlers map[string]qp.Handler
	lock     sync.Mutex
	running  uint32
	shutdown chan qp.Signal
	stopChan chan stop.Signal
	logger   qp.Logger
}

// ensure the interface is satisfied
var _ qp.PubSubTransport = (*PubSub)(nil)

// NewPubSub makes a new PubSub redis transport.
func NewPubSub(url string) *PubSub {
	return NewPubSubTimeout(url, 1*time.Second, 1*time.Second, 1*time.Second)
}

// NewPubSubTimeout makes a new PubSub redis transport and allows you to specify timeout values.
func NewPubSubTimeout(url string, connectTimeout, readTimeout, writeTimeout time.Duration) *PubSub {
	if readTimeout == 0 {
		readTimeout = 1 * time.Second
	}
	var pool = &redis.Pool{
		MaxIdle:     8,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", url, connectTimeout, readTimeout, writeTimeout)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
	p := &PubSub{
		pool:     pool,
		handlers: make(map[string]qp.Handler),
		shutdown: make(chan qp.Signal),
		stopChan: stop.Make(),
		logger:   qp.NilLogger,
	}
	return p
}

// SetLogger sets the logger to log to.
func (p *PubSub) SetLogger(logger qp.Logger) {
	p.logger = logger
}

// Publish publishes data on the specified channel.
func (p *PubSub) Publish(channel string, data []byte) error {
	if atomic.LoadUint32(&p.running) == 0 {
		return qp.ErrNotRunning
	}
	conn := p.pool.Get()
	_, err := conn.Do("PUBLISH", channel, data)
	conn.Close()
	return err
}

// Subscribe binds the handler to the specified channel.
func (p *PubSub) Subscribe(channel string, handler qp.Handler) error {
	if atomic.LoadUint32(&p.running) == 1 {
		return qp.ErrRunning
	}
	p.lock.Lock()
	p.handlers[channel] = handler
	p.lock.Unlock()
	return nil
}

func (p *PubSub) processMessages() {
	go func() {
		for c, h := range p.handlers {
			go func(channel string, handler qp.Handler) {
				conn := p.pool.Get()
				psc := redis.PubSubConn{Conn: conn}
				psc.PSubscribe(channel)
				for {
					select {
					case <-p.shutdown:
						psc.PUnsubscribe(channel)
						conn.Close()
						return
					default:
						switch v := psc.Receive().(type) {
						case redis.PMessage:
							go handler.Handle(&qp.Message{Source: v.Channel, Data: v.Data})
						case error:
							// Network timeout is fine also.
							if netErr, ok := v.(net.Error); ok {
								if netErr.Timeout() {
									continue
								}
							}
							p.logger.Error("Error when receiving from Redis:", v)
						}
					}
				}
			}(c, h)
		}
	}()
}

// Start starts the transport.
func (p *PubSub) Start() error {
	// TODO: discuss blocking this until all subscription
	// acks are received?
	if atomic.LoadUint32(&p.running) == 0 {
		atomic.StoreUint32(&p.running, 1)
		p.processMessages()
	} else {
		return qp.ErrRunning
	}
	return nil
}

// Stop stops the transport and closes StopChan() when finished.
func (p *PubSub) Stop(grace time.Duration) {
	// stop processing new Publish calls
	atomic.StoreUint32(&p.running, 0)
	// wait for duration to allow in-flight requests to finish
	time.Sleep(grace)
	// instruct all listening goroutines to shutdown
	close(p.shutdown)
	// inform caller of stop complete
	close(p.stopChan)
}

// StopChan gets the stop channel which will be closed when
// this transport has successfully stopped.
func (p *PubSub) StopChan() <-chan stop.Signal {
	return p.stopChan
}
