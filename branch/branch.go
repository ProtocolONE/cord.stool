package branch

import (
	"cord.stool/cordapi"
	"cord.stool/service/models"
	"fmt"
	"time"
)

func CreateBranch(url string, login string, password string, name string, gameID string) error {

	fmt.Printf("Creating branch ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.CreateBranch(&models.Branch{"", name, gameID, "", time.Time{}, time.Time{}})
	if err != nil {
		return err
	}

	fmt.Printf("Branch \"%s\" with id %s is created\n", result.Name, result.ID)
	return nil
}

func DeleteBranch(url string, login string, password string, id string, name string, gameID string) error {

	fmt.Printf("Deleting branch ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.DeleteBranch(id, name, gameID)
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

	result, err := api.ListBranch(gameID)
	if err != nil {
		return err
	}

	fmt.Printf("\n")
	fmt.Printf("|          APPLICATION ID          |            BRANCH ID             |               NAME               |           LIVE BUILD ID          |         CREATED AT        |\n")
	fmt.Printf("| -------------------------------- | -------------------------------- | -------------------------------- | -------------------------------- | ------------------------- |\n")
	for _, b := range result.List {
		fmt.Printf("| %32s | %32s | %32s | %32s | %24s |\n", b.GameID, b.ID, b.Name, b.BuildID, b.Created.Format("2006-01-02 15:04:05 -0700"))
	}
	fmt.Printf("\n")

	return nil
}

func ShallowBranch(url string, login string, password string, sid string, sname string, tid string, tname string, gameID string) error {

	fmt.Printf("Shallowing branch ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.ShallowBranch(sid, sname, tid, tname, gameID)
	if err != nil {
		return err
	}

	fmt.Printf("Branch \"%s\" with id %s is shallowed with \"%s\" with id %s\n", result.SourceName, result.SourceID, result.TargetName, result.TargetID)
	return nil
}
