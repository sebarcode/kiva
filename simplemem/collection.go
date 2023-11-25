package simplemem

import (
	"context"
	"sync"
	"time"

	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
)

type collection struct {
	mtx      *sync.RWMutex
	keys     []string                     // store keys, requres in order to return keys in ordered format as of inserted
	items    codekit.M                    // store item in map format
	metas    map[string]kiva.ItemMetadata // store metadata
	opts     *kiva.CacheOptions
	cancelFn context.CancelFunc
}

func (c *collection) Close() {
	if c.cancelFn != nil {
		c.cancelFn()
	}
}

func (c *collection) setOpts(opts kiva.CacheOptions) {
	if c.cancelFn != nil {
		c.cancelFn()
	}

	c.opts = &opts
	ctx, fn := context.WithCancel(context.Background())
	c.cancelFn = fn
	go c.sync(ctx)
}

func (c *collection) sync(ctx context.Context) {
	if c.opts == nil {
		return
	}

	if c.opts.ExpiryBy == kiva.ExpiryByType("") {
		return
	}

	if c.opts.ExpiryPeriod <= 0 || c.opts.SyncEvery == 0 {
		return
	}

	isSyncing := false
loopTime:
	for {
		select {
		case <-time.After(c.opts.SyncEvery):
			if isSyncing {
				continue loopTime
			}

			isSyncing = true
			newKeys := []string{}
			keys := c.keys
		keyLoop:
			for _, key := range keys {
				meta, ok := c.metas[key]
				if !ok {
					continue keyLoop
				}

				if !meta.IsExpired(*c.opts) {
					newKeys = append(newKeys, key)
					continue keyLoop
				}

				delete(c.items, key)
				delete(c.metas, key)
			}
			c.keys = newKeys
			isSyncing = false

		case <-ctx.Done():
			return
		}
	}
}
