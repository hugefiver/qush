package main

import (
	"io"
	"sync"
)

func PipeChannels(c1, c2 io.ReadWriteCloser) {
	var once sync.Once
	closeFun := func() {
		c1.Close()
		c2.Close()
	}

	// Pipe session to bash and visa-versa
	go func() {
		io.Copy(c1, c2)
		once.Do(closeFun)
	}()

	go func() {
		io.Copy(c2, c1)
		once.Do(closeFun)
	}()
}
