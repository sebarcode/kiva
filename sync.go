package kiva

import (
	"context"
	"time"
)

func (k *Kiva) SetContext(ctx context.Context) {
	k.ctx = ctx
}

func (kv *Kiva) Sync() {
	for {
		select {
		case <-kv.ctx.Done():
			return

		case <-time.After(time.Duration(kv.opts.SyncBatch.EveryInSecond) * time.Second):
			keys := kv.provider.Keys("*")
			for _, key := range keys {
				item := kv.provider.Get(key)
				if item.Error != nil {
					kv.provider.Delete(key)
				}
			}
		}
	}
}
