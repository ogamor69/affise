package app

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

type LimitedListener interface {
	Accept() (net.Conn, error)
	Close() error
	Addr() net.Addr
	Wait()
}

type LimitedListenerConfig struct {
	Listener      net.Listener
	Limit         int
	AcceptTimeout time.Duration
	CloseTimeout  time.Duration
	ErrorCallback func(error)
}

type limitedListenerImpl struct {
	net.Listener
	sem           chan struct{}
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	errorCallback func(error)
}

func NewLimitedListener(config LimitedListenerConfig) LimitedListener {
	ctx, cancel := context.WithCancel(context.Background())
	return &limitedListenerImpl{
		Listener:      config.Listener,
		sem:           make(chan struct{}, config.Limit),
		ctx:           ctx,
		cancel:        cancel,
		errorCallback: config.ErrorCallback,
	}
}

func (l *limitedListenerImpl) Addr() net.Addr {
	return l.Listener.Addr()
}

func (l *limitedListenerImpl) Accept() (net.Conn, error) {
	select {
	case l.sem <- struct{}{}:
		l.wg.Add(1)
	case <-l.ctx.Done():
		return nil, errors.New("listener closed")
	}

	conn, err := l.Listener.Accept()
	if err != nil {
		<-l.sem
		l.wg.Done()
		if l.errorCallback != nil {
			l.errorCallback(err)
		}
		return nil, err
	}
	return &limitedConn{Conn: conn, listener: l}, nil
}

func (l *limitedListenerImpl) Close() error {
	l.cancel()
	return l.Listener.Close()
}

func (l *limitedListenerImpl) Wait() {
	l.wg.Wait()
}

type limitedConn struct {
	net.Conn
	listener *limitedListenerImpl
}

func (c *limitedConn) Close() error {
	err := c.Conn.Close()
	<-c.listener.sem
	c.listener.wg.Done()
	if err != nil && c.listener.errorCallback != nil {
		c.listener.errorCallback(err)
	}
	return err
}
