package games

import (
	"cord.stool/cordapi"
	"cord.stool/qilinapi"
	"cord.stool/service/models"
	"fmt"
)

func ListGame(qilinUrl string, url string, login string, password string) error {

	fmt.Printf("Getting game list ...\n")

	api, err := qilinapi.NewQilinAPI(qilinUrl)
	if err != nil {
		return err
	}

	vendors, err := api.ListVendor()
	if err != nil {
		return err
	}

	if vendors == nil {
		return fmt.Errorf("No venders were found")
	}

	var gameList []*models.GameInfo

	for _, v := range *vendors {

		games, err := api.ListGame(v.ID)
		if err != nil {
			return err
		}

		if games == nil {
			continue
		}

		for _, g := range *games {

			info, err := api.GetGameInfo(g.ID)
			if err != nil {
				return err
			}

			if info == nil {
				continue
			}

			gameList = append(gameList, info)
		}
	}

	if len(gameList) == 0 {
		return fmt.Errorf("There are no one game found")
	}

	api2 := cordapi.NewCordAPI(url)
	err = api2.Login(login, password)
	if err != nil {
		return err
	}

	fmt.Printf("\n")
	fmt.Printf("|            APPLICATION ID            |                        NAME                        |           LIVE BUILD ID          |            LIVE BRANCH           |         RELEASED AT       |\n")
	fmt.Printf("| ------------------------------------ | -------------------------------------------------- | -------------------------------- | -------------------------------- | ------------------------- |\n")
	for _, g := range gameList {

		title := g.Title
		if title == "" {
			title = g.InternalName
		}

		var liveID, branchName string
		b, _ := api2.GetLiveBranch(g.ID)
		if b != nil {

			liveID = b.LiveBuild
			branchName = b.Name
		}

		fmt.Printf("| %36s | %50s | %32s | %32s | %24s |\n", g.ID, title, liveID, branchName, g.ReleaseDate.Format("2006-01-02 15:04:05 -0700"))
	}
	fmt.Printf("\n")

	return nil
}
