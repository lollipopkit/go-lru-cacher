package cacher

type partedCacher struct {
	// 逻辑：
	// 先把active写满，再写lazy
	// active写满，将active的最近、使用最多的缓存项“移动”到lazy
	// active写满，lazy写满，删除lazy中最早、使用最少的缓存项，移动active的最近、使用最多的缓存项到lazy

	active *cacher // 最近的，“删”时优先处理该部分
	lazy   *cacher // 低优先级的，“查”时优先处理该部分
}

func NewPartedCacher(maxLength int) *partedCacher {
	r, l := _calcMaxLength(maxLength)

	return &partedCacher{
		active: NewCacher(r),
		lazy:   NewCacher(l),
	}
}

func (c *partedCacher) Set(key, value any) {
	if (c.lazy.caches[key] != nil) {
		c.lazy.Set(key, value)
		return
	}
	if !c.active.IsFull() {
		c.active.Set(key, value)
		return
	}

	delKeyInActive, lTime, lTimes := c.active.Hotest()
	v, ok := c.active.Get(delKeyInActive)
	if !c.lazy.IsFull() {
		if ok {
			c.lazy.Set(delKeyInActive, v)
			c.active.Delete(delKeyInActive)
			c.Set(key, value)
		} else {
			c.active.Set(key, value)
		}
		return
	}

	// lazy满：
	// 1、删除lazy中最早、使用最少的缓存项
	// 2、移动active的最近、使用最多的缓存项到lazy
	// 3、将新增项目添加到active
	delKeyInLazy, eTime, eTimes := c.lazy.Coldest()
	if eTime <= lTime || eTimes <= lTimes { // FIFO：先进先出，所以包含等于
		c.lazy.Delete(delKeyInLazy)
		c.lazy.Set(delKeyInActive, v)
		c.active.Delete(delKeyInActive)
		c.Set(key, value)
		return
	}

	// lazy满，但lazy中最早、使用最少的缓存项不是active中最近、使用最多的缓存项
	// 所以，只需将新增项目添加到active
	c.active.Set(key, value)
}

func (c *partedCacher) Get(key any) (any, bool) {
	v, ok := c.lazy.Get(key)
	if ok {
		return v, ok
	}

	return c.active.Get(key)
}

func (c *partedCacher) Delete(key any) {
	c.active.Delete(key)
	c.lazy.Delete(key)
}

func (c *partedCacher) IsFull() bool {
	return c.active.IsFull() && c.lazy.IsFull()
}

func (c *partedCacher) Clear() {
	c.active.Clear()
	c.lazy.Clear()
}

func (c *partedCacher) Len() int {
	return c.active.Len() + c.lazy.Len()
}

func (c *partedCacher) Keys() []any {
	return append(c.active.Keys(), c.lazy.Keys()...)
}

func (c *partedCacher) Values() []any {
	return append(c.active.Values(), c.lazy.Values()...)
}

func (c *partedCacher) Map() map[any]any {
	aMap := c.active.Map()
	lMap := c.lazy.Map()
	for k := range lMap {
		aMap[k] = lMap[k]
	}
	return aMap
}
