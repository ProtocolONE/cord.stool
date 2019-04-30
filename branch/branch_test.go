package branch

import (
	"log"
	"testing"
)

func TestBranch(t *testing.T) {

	url := "http://127.0.0.1:5001"
	login := "admin"
	password := "123456"

	{
		delete1 := func() {
			err := DeleteBranch(url, login, password, "", "Test1", "0000000000")
			if err != nil {
				log.Panic(err)
			}
		}

		delete2 := func() {
			err := DeleteBranch(url, login, password, "", "Test2", "0000000000")
			if err != nil {
				log.Panic(err)
			}
		}

		err := CreateBranch(url, login, password, "Test1", "0000000000")
		if err != nil {
			log.Panic(err)
		}
		defer delete1()

		err = CreateBranch(url, login, password, "Test2", "0000000000")
		if err != nil {
			log.Panic(err)
		}
		defer delete2()

		err = ListBranch(url, login, password, "0000000000")
		if err != nil {
			log.Panic(err)
		}

		err = SetLiveBranch(url, login, password, "", "Test2", "0000000000")
		if err != nil {
			log.Panic(err)
		}

		err = ListBranch(url, login, password, "0000000000")
		if err != nil {
			log.Panic(err)
		}

		err = ShallowBranch(url, login, password, "", "Test1", "", "Test2", "0000000000")
		if err != nil {
			log.Panic(err)
		}

		err = ListBranch(url, login, password, "0000000000")
		if err != nil {
			log.Panic(err)
		}
	}

	err := ListBranch(url, login, password, "0000000000")
	if err != nil {
		log.Panic(err)
	}

	err = LiveBuild(url, login, password, "ceba80b3-60de-4fbd-9ae7-7bbfece5e5e2", "Test", "6E8D1B28C9A440BAA3E45D5E8E9032C6")
	if err != nil {
		log.Panic(err)
	}

	err = ListBuild(url, login, password, "ceba80b3-60de-4fbd-9ae7-7bbfece5e5e2", "Test")
	if err != nil {
		log.Panic(err)
	}
}
