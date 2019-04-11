package games

import (
	//"cord.stool/cordapi"
	//"cord.stool/service/models"
	"fmt"
)

func ListGame(url string, login string, password string) error {

	fmt.Printf("Getting game list ...\n")

	/*api := cordapi.NewCordAPI(url)
	err := api.Login(login, password)
	if err != nil {
		return err
	}

	result, err := api.ListBranch(&models.ListBranchCmd{gameID})
	if err != nil {
		return err
	}*/

	fmt.Printf("\n")
	fmt.Printf("|          APPLICATION ID          |               NAME               |           LIVE BUILD ID          |         CREATED AT        |\n")
	fmt.Printf("| -------------------------------- | -------------------------------- | -------------------------------- | ------------------------- |\n")
	/*for _, b := range result.List {
		fmt.Printf("| %32s | %32s | %32s | %24s |\n", b.GameID, b.ID, b.Name, b.LiveBuildID, b.Created.Format("2006-01-02 15:04:05 -0700"))
	}*/
	fmt.Printf("\n")

	return nil
}
