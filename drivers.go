package midi

import (
	"fmt"
	"sync"
)

var (
	driverLock = sync.RWMutex{}
	drivers    = map[string]Driver{}
)

func Register(d Driver) {
	driverLock.Lock()
	defer driverLock.Unlock()

	_, existing := drivers[d.Name()]
	if existing {
		panic(fmt.Sprintf("driver '%s' is already registered", d.Name()))
	}

	drivers[d.Name()] = d
}

func DriverByName(name string) Driver {
	driverLock.RLock()
	defer driverLock.RUnlock()

	drv, ok := drivers[name]
	if !ok {
		panic(fmt.Sprintf("unknown driver '%s'", name))
	}

	return drv
}
