package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jordanp/goapp/entity"
	"github.com/jordanp/goapp/pkg/log"
	"github.com/jordanp/goapp/store"
)

type User struct {
	log          log.Logger
	store        *store.User
	updateTicker *time.Ticker

	sync.RWMutex
	users map[string]entity.User
}

func NewUserCache(log log.Logger, store *store.User) (*User, error) {
	c := &User{
		log:   log,
		store: store,
		users: make(map[string]entity.User),
	}

	if err := c.updateLoop(15 * time.Second); err != nil {
		return nil, err
	}

	return c, nil
}

func (cache *User) updateLoop(updateFrequency time.Duration) error {
	errorMsg := "unable to update User cache from Store: %s"
	if err := cache.Update(); err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	cache.updateTicker = time.NewTicker(updateFrequency)
	go func() {
		for range cache.updateTicker.C {
			if err := cache.Update(); err != nil {
				cache.log.Errorf(errorMsg, err)
			}
		}
	}()

	return nil
}

func (cache *User) Stop() {
	if cache.updateTicker != nil {
		cache.updateTicker.Stop()
	}
}

func (cache *User) Update() error {
	users, err := cache.store.GetAll(context.Background())
	if err != nil {
		return err
	}

	m := make(map[string]entity.User, len(users))
	for _, user := range users {
		m[user.Login] = user
	}

	cache.Lock()
	cache.users = m
	cache.Unlock()
	return nil
}

func (cache *User) GetByLogin(login string) entity.User {
	cache.RLock()
	CDNs := cache.users[login]
	cache.RUnlock()
	return CDNs
}
