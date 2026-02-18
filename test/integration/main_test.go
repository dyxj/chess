package integration

import (
	"log"
	"os"
	"testing"

	"github.com/dyxj/chess/test/testx"
)

func TestMain(m *testing.M) {
	var code int
	defer func() {
		os.Exit(code)
	}()

	ready, errChan := testx.RunGlobalEnv()

	select {
	case <-ready:
		log.Println("global env ready")
	case eErr := <-errChan:
		log.Panicf("failed to start global env: %v", eErr)
	}
	defer testx.GlobalEnv().Close()

	code = m.Run()
}
