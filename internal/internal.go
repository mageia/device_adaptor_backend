package internal

import (
	"bytes"
	"context"
	"crypto/rand"
	"math/big"
	"strconv"
	"time"
)

type Duration struct {
	Duration time.Duration
}

func (d *Duration) UnmarshalTOML(b []byte) error {
	var err error
	b = bytes.Trim(b, `'`)
	if d.Duration, err = time.ParseDuration(string(b)); err == nil {
		return nil
	}

	if uq, err := strconv.Unquote(string(b)); err == nil && len(uq) > 0 {
		if d.Duration, err = time.ParseDuration(uq); err == nil {
			return nil
		}
	}

	if sI, err := strconv.ParseInt(string(b), 10, 64); err == nil {
		d.Duration = time.Second * time.Duration(sI)
		return nil
	}

	if sF, err := strconv.ParseFloat(string(b), 64); err == nil {
		d.Duration = time.Second * time.Duration(sF)
		return nil
	}
	return nil
}

func RandomSleep(max time.Duration, ctx context.Context) {
	if max == 0 {
		return
	}
	maxSleep := big.NewInt(max.Nanoseconds())
	var sleepNs int64
	if j, err := rand.Int(rand.Reader, maxSleep); err == nil {
		sleepNs = j.Int64()
	}
	t := time.NewTimer(time.Nanosecond * time.Duration(sleepNs))
	select {
	case <-t.C:
		return
	case <-ctx.Done():
		t.Stop()
		return
	}
}
