package app

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestLimitedListenerImpl(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	limitedListener := NewLimitedListener(LimitedListenerConfig{
		Listener:      listener,
		Limit:         100,
		AcceptTimeout: time.Second * 5,
	})
	defer limitedListener.Close()

	go func() {
		for {
			conn, err := limitedListener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	clients := 5
	var wg sync.WaitGroup
	wg.Add(clients)

	for i := 0; i < clients; i++ {
		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", listener.Addr().String())
			if err != nil {
				t.Logf("Error dialing: %v", err)
				return
			}
			conn.Close()
		}()
	}

	wg.Wait()
	limitedListener.Close()
	limitedListener.Wait()
}
