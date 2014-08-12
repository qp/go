package redis

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/qp/go"
	"github.com/stretchr/pat/stop"
	"github.com/stretchr/slog"
)

// PubSub represents a qp.PubSubTransport.
type PubSub struct {
	pool     *redis.Pool
	handlers map[string]qp.Handler
	lock     sync.Mutex
	running  uint32
	shutdown chan qp.Signal
	stopChan chan stop.Signal
	log      slog.Logger
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
		log:      slog.NilLogger,
	}
	return p
}

// SetLogger sets the Logger to log to.
func (p *PubSub) SetLogger(log slog.Logger) {
	p.log = log
}

// Publish publishes data on the specified channel.
func (p *PubSub) Publish(channel string, data []byte) error {
	if atomic.LoadUint32(&p.running) == 0 {
		return qp.ErrNotRunning
	}
	if p.log.Info() {
		p.log.Info("publish to", channel, string(data))
	}
	conn := p.pool.Get()
	_, err := conn.Do("PUBLISH", channel, data)
	conn.Close()
	if err != nil && p.log.Err() {
		p.log.Err("publish failed", err)
	}
	return err
}

// Subscribe binds the handler to the specified channel.
func (p *PubSub) Subscribe(channel string, handler qp.Handler) error {
	if atomic.LoadUint32(&p.running) == 1 {
		return qp.ErrRunning
	}
	if p.log.Info() {
		p.log.Info("subscribing to", channel)
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
						if p.log.Info() {
							p.log.Info("shutting down")
						}
						psc.PUnsubscribe(channel)
						conn.Close()
						return
					default:
						switch v := psc.Receive().(type) {
						case redis.PMessage:
							if p.log.Info() {
								p.log.Info("handling message from", v.Channel, v.Data)
							}
							go handler.Handle(&qp.Message{Source: v.Channel, Data: v.Data})
						case net.Error:
							// Network timeout is fine also.
							if v.(net.Error).Timeout() {
								if p.log.Info() {
									p.log.Info("network timeout, refreshing.")
								}
								continue
							}
							if p.log.Warn() {
								p.log.Warn("error when receiving from Redis:", v)
							}
						case error:
							if p.log.Warn() {
								// TODO: decide what's meant to happen at this point -
								// at the moment, we just get millions of error reports.
								// To recreate:
								// 1. start redis
								// 2. run a service (or other direct transporter thing)
								// 3. stop redis
								// 4. watch logs
								p.log.Warn("error when receiving from Redis:", v)
							}
							return
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
		p.log.Info("starting")
		p.processMessages()
	} else {
		return qp.ErrRunning
	}
	return nil
}

// Stop stops the transport and closes StopChan() when finished.
func (p *PubSub) Stop(grace time.Duration) {
	p.log.Info("stopping...")
	// stop processing new Publish calls
	atomic.StoreUint32(&p.running, 0)
	// wait for duration to allow in-flight requests to finish
	time.Sleep(grace)
	// instruct all listening goroutines to shutdown
	close(p.shutdown)
	// inform caller of stop complete
	close(p.stopChan)
	p.log.Info("stopped")
}

// StopChan gets the stop channel which will be closed when
// this transport has successfully stopped.
func (p *PubSub) StopChan() <-chan stop.Signal {
	return p.stopChan
}
