package kiva

import (
	"context"
	"io"
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
				item := map[string]interface{}{}
				opt, err := kv.provider.Get(key, &item)
				if err == nil {
					// data exist on hs
					switch opt.SyncDirection {
					case SyncToHots:
						if kv.getter == nil {
							break
						}
						newItem := map[string]interface{}{}
						getterErr := kv.getter(key, "", GetByID, &newItem)
						if getterErr == io.EOF {
							kv.provider.Delete(key)
							break
						} else if getterErr != nil {
							break
						}
						item = newItem
						kv.Set(key, item, &kv.opts.DefaultWrite, false)

					case SyncToPersistent:
						if kv.commiter == nil {
							break
						}
						err := kv.commiter(key, item, CommitSave)
						if err != nil {
							break
						}
						opt.SyncDirection = SyncToHots
						kv.provider.ChangeSyncOpts(key, opt)
					}

				} else {
					// data not exist on hs, then delete it from hs
					kv.provider.Delete(key)
				}
			}
		}
	}
}
