package cacher

import "time"

// 返回一个缓存器，每经过 duration 调用一次 fn 自定义清理缓存项
func NewDurationCacher(maxLength int, checkDuration time.Duration, fn func(key any, item *CacheItem) bool) *cacher {
	c := NewCacher(maxLength)
	for range time.Tick(checkDuration) {
		go c.DeleteAll(fn)
	}
	return c
}

// 每过 checkDuration 检查一次，间隔超过 elapsedDuration 的缓存项将被清理
func NewElapsedCacher(maxLength int, checkDuration, elapsedDuration time.Duration) *cacher {
	elapsedUnixNano := elapsedDuration.Nanoseconds()
	c := NewDurationCacher(maxLength, checkDuration, func(key any, item *CacheItem) bool {
		return _unixNano()-item.LastTime > elapsedUnixNano
	})
	return c
}
