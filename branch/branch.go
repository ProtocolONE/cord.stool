package branch

import (
	"cord.stool/cordapi"
	"cord.stool/service/models"
	"fmt"
)

func CreateBranch(url string, login string, password string, gameID string, name string) error {

	fmt.Printf("Creating branch ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.CreateBranch(&models.BranchInfoCmd{"", name, gameID})
	if err != nil {
		return err
	}

	fmt.Printf("Branch \"%s\" with id %s is created\n", result.Name, result.ID)
	return nil
}

func DeleteBranch(url string, login string, password string, gameID string, nameOrID string) error {

	fmt.Printf("Deleting branch ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.DeleteBranch(&models.BranchInfoCmd{nameOrID, nameOrID, gameID})
	if err != nil {
		return err
	}

	fmt.Printf("Branch \"%s\" with id %s is deleted\n", result.Name, result.ID)
	return nil
}

func ListBranch(url string, login string, password string, gameID string) error {

	fmt.Printf("Getting branch list ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.ListBranch(&models.ListBranchCmd{gameID})
	if err != nil {
		return err
	}

	fmt.Printf("\n")
	fmt.Printf("|          APPLICATION ID          |            BRANCH ID             |               NAME               |           LIVE BUILD ID          |         CREATED AT        |\n")
	fmt.Printf("| -------------------------------- | -------------------------------- | -------------------------------- | -------------------------------- | ------------------------- |\n")
	for _, b := range result.List {
		fmt.Printf("| %32s | %32s | %32s | %32s | %24s |\n", b.GameID, b.ID, b.Name, b.LiveBuildID, b.Created.Format("2006-01-02 15:04:05 -0700"))
	}
	fmt.Printf("\n")

	return nil
}

func ShallowBranch(url string, login string, password string, sNameOrID string, tNameOrID string) error {

	fmt.Printf("Shallowing branch ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.ShallowBranch(&models.ShallowBranchCmd{sNameOrID, tNameOrID})
	if err != nil {
		return err
	}

	fmt.Printf("Branch \"%s\" with id %s is shallowed with \"%s\" with id %s\n", result.SourceName, result.SourceID, result.TargetName, result.TargetID)
	return nil
}
