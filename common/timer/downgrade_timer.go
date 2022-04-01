package timer

import (
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type DowngradeTimer struct {
	min      time.Duration
	max      time.Duration
	max_idle int64

	cur      time.Duration
	cur_idle int64

	urgent_chan chan interface{}
}

func NewDowngradeTimer(min time.Duration, max time.Duration, idle int64) *DowngradeTimer {
	return &DowngradeTimer{
		cur:      min,
		min:      min,
		max:      max,
		cur_idle: 0,
		max_idle: idle,

		urgent_chan: make(chan interface{}, 100),
	}
}

func (c *DowngradeTimer) Wait() {
	idle_cnt := int64(0)
	for {
		cur_idle := atomic.LoadInt64(&c.cur_idle)
		if !atomic.CompareAndSwapInt64(&c.cur_idle, cur_idle, cur_idle+1) {
			continue
		}
		idle_cnt = cur_idle
		break
	}

	if idle_cnt >= c.max_idle {
		c.cur = c.cur * 2
		if c.cur > c.max {
			c.cur = c.max
		}
		logrus.Debugf("cur idle %d, increse timeout to %s", idle_cnt, c.cur)
	}

	t := time.NewTimer(c.cur)

	select {
	case <-c.urgent_chan:
		c.cur = c.min
	case <-t.C:
	}
}

func (c *DowngradeTimer) Trigger() {
	cur_idle := atomic.LoadInt64(&c.cur_idle)
	atomic.StoreInt64(&c.cur_idle, 0)

	if cur_idle >= c.max_idle {
		c.urgent_chan <- struct{}{}
		logrus.Infof("send urgent signal to timer")
	}
}
