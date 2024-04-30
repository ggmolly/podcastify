package shared

import (
	"path/filepath"
	"sync"

	"github.com/ggmolly/podcastify/orm"
)

var (
	mapLock   = sync.RWMutex{}
	StreamMap = make(map[string]*orm.Podcast)
)

func UpdateStream(name string, podcast *orm.Podcast) bool {
	mapLock.Lock()
	defer mapLock.Unlock()
	if _, ok := StreamMap[name]; ok {
		StreamMap[name] = podcast
		return true
	}
	return false
}

func GetStream(name string) *orm.Podcast {
	mapLock.RLock()
	defer mapLock.RUnlock()
	if podcast, ok := StreamMap[name]; ok {
		return podcast
	}
	return nil
}

func GetStreamFromPath(path string) *orm.Podcast {
	mapLock.RLock()
	defer mapLock.RUnlock()
	name := filepath.Base(path)
	return GetStream(name)
}
