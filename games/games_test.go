package games

import (
	"log"
	"testing"
)

func TestGame(t *testing.T) {

	err := ListGame("https://qilinapi.tst.protocol.one", "http://127.0.0.1:5001", "admin", "123456")
	if err != nil {
		log.Panic(err)
	}
}
