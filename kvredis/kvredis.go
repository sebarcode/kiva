package kvredis

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"

	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/serde"
	"github.com/go-redis/redis/v8"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
	"github.com/sebarcode/logger"
)

type RedisProvider struct {
	ctx   context.Context
	rdb   *redis.Client
	byter byter.Byter
}

func New(connTxt string, logger *logger.LogEngine, dataByter byter.Byter) (*RedisProvider, error) {
	p := new(RedisProvider)
	parts, err := url.Parse(connTxt)
	if err != nil {
		return nil, fmt.Errorf("connection text parse error. %s", err.Error())
	}
	//userid := parts.User.Username()
	password, _ := parts.User.Password()
	host := parts.Host
	dbnum := strings.Trim(parts.Path, "//")

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,                                   // no passwordset
		DB:       codekit.ToInt(dbnum, codekit.RoundingAuto), // use default DB
	})

	p.ctx = ctx
	p.rdb = rdb
	if dataByter == nil {
		p.byter = byter.NewByter("")
	} else {
		p.byter = dataByter
	}

	return p, nil
}

func (p *RedisProvider) Connect() error {
	if _, e := p.rdb.Ping(p.ctx).Result(); e != nil {
		return e
	}
	return nil
}

func (p *RedisProvider) Close() {
	p.rdb.Close()
}

func (p *RedisProvider) Set(key string, value interface{}, opts *kiva.WriteOptions) error {
	rv := reflect.ValueOf(value)
	kind := rv.Kind()

	var err error
	if kind == reflect.Ptr || kind == reflect.Map || kind == reflect.Slice || kind == reflect.Struct {
		var bs []byte
		if bs, err = p.byter.Encode(value); err != nil {
			return err
		}
		_, err = p.rdb.Set(p.ctx, key, bs, opts.TTL).Result()
	} else {
		_, err = p.rdb.Set(p.ctx, key, value, opts.TTL).Result()
	}
	return err
}

func (p *RedisProvider) Get(key string, dest interface{}) error {
	str, err := p.rdb.Get(p.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return io.EOF
		}
		return err
	}

	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("output should be a pointer")
	}
	kind := rv.Elem().Kind()

	// use byte for non primitive data type
	if kind == reflect.Ptr || kind == reflect.Map || kind == reflect.Struct || kind == reflect.Slice {
		if err = p.byter.DecodeTo([]byte(str), dest, nil); err != nil {
			return err
		}
		return nil
	}

	// it is a primitive data type, use serde
	if err = serde.Serde(str, dest); err != nil {
		return err
	}
	return nil
}

func (p *RedisProvider) Delete(key string) {
	p.rdb.Del(p.ctx, key).Result()
}

func (p *RedisProvider) getKeysInRange(from, to string) []string {
	pattern := ""
	for idx, c := range from {
		ct := to[idx]
		if byte(c) != ct {
			break
		}
		pattern += string(c)
	}
	pattern += "*"
	keys, _ := p.rdb.Keys(p.ctx, pattern).Result()

	inRangeKeys := []string{}
	for _, key := range keys {
		if strings.Compare(key, from) >= 0 && strings.Compare(key, to) <= 0 {
			inRangeKeys = append(inRangeKeys, key)
		}
	}
	return inRangeKeys
}

func (p *RedisProvider) Keys(pattern string) []string {
	keys, _ := p.rdb.Keys(p.ctx, pattern).Result()
	return keys
}

func (p *RedisProvider) KeyRanges(from, to string) []string {
	return p.getKeysInRange(from, to)
}
