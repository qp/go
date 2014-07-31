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

// Direct represents a qp.DirectTransport.
type Direct struct {
	pool     *redis.Pool
	stopChan chan stop.Signal
	running  uint32
	handlers map[string]qp.Handler
	lock     sync.Mutex
	shutdown chan qp.Signal
	logger   qp.Logger
}

// ensure the interface is satisfied
var _ qp.DirectTransport = (*Direct)(nil)

// NewDirect makes a new Direct redis transport.
func NewDirect(url string) *Direct {
	return NewDirectTimeout(url, 1*time.Second, 1*time.Second, 1*time.Second)
}

// NewDirectTimeout makes a new Direct redis transport and allows you to specify timeout values.
func NewDirectTimeout(url string, connectTimeout, readTimeout, writeTimeout time.Duration) *Direct {
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
	p := &Direct{
		pool:     pool,
		handlers: make(map[string]qp.Handler),
		shutdown: make(chan qp.Signal),
		stopChan: stop.Make(),
		logger:   qp.NilLogger,
	}
	return p
}

// SetLogger sets the logger to log to.
func (d *Direct) SetLogger(logger qp.Logger) {
	d.logger = logger
}

// Send sends data on the channel.
func (d *Direct) Send(channel string, data []byte) error {
	if atomic.LoadUint32(&d.running) == 0 {
		return qp.ErrNotRunning
	}
	conn := d.pool.Get()
	_, err := conn.Do("LPUSH", channel, data)
	conn.Close()
	return err
}

// OnMessage binds the handler to the specified channel.
func (d *Direct) OnMessage(channel string, handler qp.Handler) error {
	if atomic.LoadUint32(&d.running) == 1 {
		return qp.ErrRunning
	}
	d.lock.Lock()
	d.handlers[channel] = handler
	d.lock.Unlock()
	return nil
}

func (d *Direct) processMessages() {
	go func() {
		for c, h := range d.handlers {
			go func(channel string, handler qp.Handler) {
				for {
					select {
					case <-d.shutdown:
						return
					default:
						conn := d.pool.Get()
						if err := d.handleMessage(conn, channel, handler); err != nil {
							d.logger.Error("Failed to handle message:", err)
						}
						conn.Close()
					}
				}
			}(c, h)
		}
	}()
}

func (d *Direct) handleMessage(conn redis.Conn, channel string, handler qp.Handler) error {
	var data []byte
	// BRPOP on the channel to wait for a new message
	message, err := redis.Values(conn.Do("BRPOP", channel, "1"))
	if err != nil {
		// Did the BRPOP return with no data?
		if err == redis.ErrNil {
			return nil
		}
		// Network timeout is fine also.
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				return nil
			}
		}
		return err
	}
	if _, err := redis.Scan(message, &channel, &data); err != nil {
		return err
	}
	go handler.Handle(&qp.Message{Source: channel, Data: data})
	return nil
}

// Start starts the transport.
func (d *Direct) Start() error {
	if atomic.LoadUint32(&d.running) == 0 {
		atomic.StoreUint32(&d.running, 1)
		go d.processMessages()
	} else {
		return qp.ErrRunning
	}
	return nil
}

// Stop instructs the transport to gracefully stop and close the
// StopChan when stopping has completed.
//
// In-flight requests will have "wait" duration to complete
// before being abandoned.
func (d *Direct) Stop(grace time.Duration) {
	// stop processing new Sends
	atomic.StoreUint32(&d.running, 0)
	// wait for duration to allow in-flight requests to finish
	time.Sleep(grace)
	// instruct all receiving goroutines to shutdown
	close(d.shutdown)
	// inform caller of stop complete
	close(d.stopChan)
}

// StopChan gets the stop channel which will block until
// stopping has completed, at which point it is closed.
// Callers should never close the stop channel.
func (d *Direct) StopChan() <-chan stop.Signal {
	return d.stopChan
}
