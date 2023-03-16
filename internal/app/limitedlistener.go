package app

import (
	"net"
	"sync"
)

type LimitedListener struct {
	net.Listener
	sem chan struct{}
	wg  sync.WaitGroup
}

func NewLimitedListener(listener net.Listener, limit int) *LimitedListener {
	return &LimitedListener{
		Listener: listener,
		sem:      make(chan struct{}, limit),
	}
}

func (l *LimitedListener) Accept() (net.Conn, error) {
	l.sem <- struct{}{}
	l.wg.Add(1)

	conn, err := l.Listener.Accept()
	if err != nil {
		<-l.sem
		l.wg.Done()
	}
	return &limitedConn{Conn: conn, listener: l}, err
}

func (l *LimitedListener) Close() error {
	return l.Listener.Close()
}

func (l *LimitedListener) Wait() {
	l.wg.Wait()
}

type limitedConn struct {
	net.Conn
	listener *LimitedListener
}

func (c *limitedConn) Close() error {
	err := c.Conn.Close()
	<-c.listener.sem
	c.listener.wg.Done()
	return err
}
