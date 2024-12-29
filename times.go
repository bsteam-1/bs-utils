package utils

import (
    "sync"
    "time"

    "github.com/beevik/ntp"
)

type NowTime struct {
    t  time.Time
    mu sync.Mutex
}

func (t *NowTime) Time() time.Time {
    t.mu.Lock()
    defer t.mu.Unlock()
    return t.t
}

func (t *NowTime) SetTime() {
    t.mu.Lock()
    t.t, _ = ntp.Time("time.windows.com")
    t.mu.Unlock()
}
