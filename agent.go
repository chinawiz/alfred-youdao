package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/zgs225/youdao"
)

const (
	CACHE_EXPIRES      time.Duration = 30 * 24 * time.Hour
	CACHE_FILE         string        = "cache.dat"
	CACHE_HISTORY_FILE string        = "history.dat" // Local cache file for history data
)

type agentClient struct {
	Client  *youdao.Client
	Cache   *cache.Cache
	History *cache.Cache
	Dirty   bool
}

func (a *agentClient) Query(q string) (*youdao.Result, error) {
	k := fmt.Sprintf("from:%s,to:%s,q:%s", a.Client.GetFrom(), a.Client.GetTo(), q)
	v, ok := a.Cache.Get(k)
	if ok {
		log.Println("Cache hit")
		return v.(*youdao.Result), nil
	}
	log.Println("Cache miss")
	r, err := a.Client.Query(q)
	if err != nil {
		return nil, err
	}
	a.Cache.Set(k, r, CACHE_EXPIRES)
	a.Dirty = true
	return r, nil
}

func newAgent(c *youdao.Client) *agentClient {
	gob.Register(&youdao.Result{})
	c2 := cache.New(CACHE_EXPIRES, CACHE_EXPIRES)
	c3 := cache.New(CACHE_EXPIRES, CACHE_EXPIRES) // new
	err := c2.LoadFile(CACHE_FILE)
	if err != nil {
		log.Println(err)
	}
	err = c3.LoadFile(CACHE_HISTORY_FILE) // new
	if err != nil {
		log.Println(err)
	}
	log.Println("Cache count:", c2.ItemCount())
	log.Println("History count:", c3.ItemCount()) // new
	return &agentClient{c, c2, c3, false}
}

// Write the history data
func (a *agentClient) writeHistory(queue []string) error {
	for i, v := range queue {
		a.History.Set(strconv.Itoa(i), v, CACHE_EXPIRES)
		log.Println(strconv.Itoa(i), v)
	}
	a.Dirty = true
	return nil
}

// Retrieve the history data
func (a *agentClient) readHistory() ([]string, error) {
	queue := make([]string, 0, QUEUE_SIZE)
	for i := 0; i < QUEUE_SIZE; i++ {
		v, ok := a.History.Get(strconv.Itoa(i))
		if ok {
			log.Println("History Cache hit", strconv.Itoa(i), v)
			queue = append(queue, v.(string))
		} else {
			log.Println("History Cache miss")
		}
	}
	return queue, nil
}
