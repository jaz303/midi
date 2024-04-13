package midi

import (
	"fmt"
	"sync"
)

var (
	driverLock = sync.RWMutex{}
	drivers    = map[string]*Stub{}
)

type Stub struct {
	Name         string
	Available    bool
	CreateDriver func(name string) (Driver, error)
}

func Register(d *Stub) {
	driverLock.Lock()
	defer driverLock.Unlock()

	_, existing := drivers[d.Name]
	if existing {
		panic(fmt.Sprintf("driver '%s' is already registered", d.Name))
	}

	drivers[d.Name] = d
}

func NewDriverByName(driverName string, clientName string) (Driver, error) {
	driverLock.RLock()
	defer driverLock.RUnlock()

	drv, ok := drivers[driverName]
	if !ok {
		panic(fmt.Sprintf("unknown driver '%s'", driverName))
	}

	return drv.CreateDriver(clientName)
}
