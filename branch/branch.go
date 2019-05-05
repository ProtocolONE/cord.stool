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

	result, err := api.CreateBranch(&models.Branch{"", name, gameID, "", false, time.Time{}, time.Time{}})
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

func SetLiveBranch(url string, login string, password string, id string, name string, gameID string) error {

	fmt.Printf("Marking branch as live ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.SetLiveBranch(id, name, gameID)
	if err != nil {
		return err
	}

	fmt.Printf("Branch \"%s\" with id %s is live\n", result.Name, result.ID)
	return nil
}

func ListBranch(url string, login string, password string, gameID string) error {

	fmt.Printf("Getting branch list ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	list, err := api.ListBranch(gameID)
	if err != nil {
		return err
	}

	if list != nil {

		fmt.Printf("\n")
		fmt.Printf("|            APPLICATION ID            |            BRANCH ID             |               NAME               |              BUILD ID            |  LIVE  |         CREATED AT        |\n")
		fmt.Printf("| ------------------------------------ | -------------------------------- | -------------------------------- | -------------------------------- | ------ | ------------------------- |\n")
		for _, b := range *list {
			fmt.Printf("| %36s | %32s | %32s | %32s | %6t | %24s |\n", b.GameID, b.ID, b.Name, b.LiveBuild, b.Live, b.Created.Format("2006-01-02 15:04:05 -0700"))
		}
		fmt.Printf("\n")
	} else {

		fmt.Println("There are no one branch found")

	}

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

func LiveBuild(url string, login string, password string, gameID string, branch string, buildId string) error {

	fmt.Printf("Living build ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	brn, err := api.GetBranch("", branch, gameID)
	if err != nil {
		return err
	}

	build, err := api.GetBuild(buildId)
	if err != nil {
		return err
	}

	brn.LiveBuild = build.ID
	err = api.UpdateBranch(brn)
	if err != nil {
		return err
	}

	fmt.Printf("Branch \"%s\" with id %s has live build with id %s\n", brn.Name, brn.ID, brn.LiveBuild)
	return nil
}

func ListBuild(url string, login string, password string, gameID string, branch string) error {

	fmt.Printf("Getting build list ...\n")

	api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	brn, err := api.GetBranch("", branch, gameID)
	if err != nil {
		return err
	}

	list, err := api.ListBuild(gameID, branch)
	if err != nil {
		return err
	}

	if list != nil {

		fmt.Printf("\n")
		fmt.Printf("|            APPLICATION ID            |            BRANCH ID             |               NAME               |              BUILD ID            |  LIVE  |         CREATED AT        |\n")
		fmt.Printf("| ------------------------------------ | -------------------------------- | -------------------------------- | -------------------------------- | ------ | ------------------------- |\n")
		for _, b := range *list {
			fmt.Printf("| %36s | %32s | %32s | %32s | %6t | %24s |\n", brn.GameID, brn.ID, brn.Name, b.ID, brn.LiveBuild == b.ID, b.Created.Format("2006-01-02 15:04:05 -0700"))
		}
		fmt.Printf("\n")
	} else {

		fmt.Println("There are no one build found")

	}

	return nil
}
