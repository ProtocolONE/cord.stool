package branch

import (
	"cord.stool/cordapi"
	"cord.stool/service/models"
	"fmt"
)

func CreateBranch(url string, login string, password string, gameID string, name string) error {

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.CreateBranch(&models.BranchInfoCmd{"", name, gameID})
	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}

func DeleteBranch(url string, login string, password string, gameID string, nameOrID string) error {

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.DeleteBranch(&models.BranchInfoCmd{nameOrID, nameOrID, gameID})
	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}

func ListBranch(url string, login string, password string, gameID string) error {

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.ListBranch(&models.ListBranchCmd{gameID})
	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}

func ShallowBranch(url string, login string, password string, sNameOrID string, tNameOrID string) error {

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.ShallowBranch(&models.ShallowBranchCmd{sNameOrID, tNameOrID})
	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}
