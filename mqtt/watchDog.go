package mqtt

import (
	"time"
	"github.com/sirupsen/logrus"
	"os"
)

type watchDog struct {
	timeout time.Duration
	ping    chan struct{}
	stop    chan struct{}
}

var wdLogger = logrus.WithField("where", "watchDog")

// NewWatchDog creates a new watch dog
// timeout - after this amount of time without a ping, the program will be killed and returns 3
func NewWatchDog(timeout time.Duration) *watchDog {
	wd := watchDog{
		timeout: timeout,
		ping:    make(chan struct{}),
		stop:    make(chan struct{}),
	}

	go wd.loop()

	return &wd
}

// Ping updates the watch dog timeout
func (wd *watchDog) Ping() {
	wd.ping <- struct{}{}
}

func (wd *watchDog) Stop() {
	wdLogger.Println("Stopping watch dog")
	wd.stop <- struct{}{}
}

func (wd *watchDog) loop() {
	timer := time.NewTicker(60 * time.Second)
	lastKeepAlive := time.Now()

	for {
		select {
		case <-timer.C:
			if time.Since(lastKeepAlive) > wd.timeout {
				wdLogger.Errorf("Last keep alive (%s) is older than allowed timeout (%s). Exit!",
					lastKeepAlive, wd.timeout)
				os.Exit(3)
			}
		case <-wd.ping:
			lastKeepAlive = time.Now()
		case <-wd.stop:
			timer.Stop()
			return
		}
	}
}
