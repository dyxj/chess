package testx

import (
	"log"
	"sync"
)

var (
	global     *Environment
	globalOnce sync.Once
)

func RunGlobalEnv() (<-chan struct{}, <-chan error) {
	globalOnce.Do(func() {
		global = NewEnvironment("global")
	})
	return global.Run()
}

func GlobalEnv() *Environment {
	if global == nil {
		log.Panicf("global test environment is not initialized")
	}
	return global
}
