package parser

import (
	"sync"
	"time"
)

// globalTZCache — глобальный потокобезопасный кэш часовых поясов.
// Предотвращает повторные вызовы time.LoadLocation для одних и тех же TZID.
var globalTZCache = &tzCache{
	zones: make(map[string]*time.Location),
}

// tzCache — кэш *time.Location по строковому TZID.
// Потокобезопасен через sync.RWMutex.
type tzCache struct {
	mu    sync.RWMutex
	zones map[string]*time.Location
}

// load возвращает *time.Location для указанного TZID.
// При первом обращении загружает через time.LoadLocation и кэширует.
func (c *tzCache) load(tzid string) (*time.Location, error) {
	// Быстрый путь: чтение из кэша.
	c.mu.RLock()
	loc, ok := c.zones[tzid]
	c.mu.RUnlock()
	if ok {
		return loc, nil
	}

	// Медленный путь: загрузка и запись в кэш.
	loc, err := time.LoadLocation(tzid)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.zones[tzid] = loc
	c.mu.Unlock()

	return loc, nil
}
