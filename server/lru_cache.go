package main

import (
	"container/list"
	"sync"
)

type entry struct {
	key   string
	value string
}

type LRUCache struct {
	capacity int
	list     *list.List
	items    map[string]*list.Element
	mu       sync.Mutex
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		capacity: cap,
		list:     list.New(),
		items:    make(map[string]*list.Element),
	}
}

func (c *LRUCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, found := c.items[key]; found {
		c.list.MoveToFront(element)
		return element.Value.(*entry).value, true
	}
	return "", false
}

func (c *LRUCache) Put(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, found := c.items[key]; found {
		element.Value.(*entry).value = value
		c.list.MoveToFront(element)
		return
	}

	element := c.list.PushFront(&entry{key, value})
	c.items[key] = element

	if c.list.Len() > c.capacity {
		last := c.list.Back()
		c.list.Remove(last)
		delete(c.items, last.Value.(*entry).key)
	}
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.items[key]; ok {
		c.list.Remove(ele)
		delete(c.items, key)
	}
}

