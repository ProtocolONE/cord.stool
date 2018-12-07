package upgrade

import (
	"fmt"
	"path/filepath"
	"os"
	"io"
	"io/ioutil"
	"net/http"
	"archive/zip"
	context2 "context"
	"cord.stool/context"
	"github.com/urfave/cli"
	"github.com/google/go-github/github"
)

var args = struct {
	Force bool
	Check bool
}{
	Force: false,
	Check: false,
}

func Register(ctx *context.StoolContext) {

	cmd := cli.Command{
		Name:        "upgrade",
		ShortName:   "u",
		Usage:       "Looking for upgrades",
		Description: "Upgrades application to the latest version",

		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:        "force, f",
				Usage:       "Force to upgrade self-built version",
				Destination: &args.Force,
			},
			cli.BoolFlag{
				Name:        "check, c",
				Usage:       "Checking for new version",
				Destination: &args.Check,
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func do(ctx *context.StoolContext, c *cli.Context) error {

	client := github.NewClient(nil)
	release, err := getLatestRelease(client)
	if err != nil {
		return err
	}

	needUpgrade := ctx.Version != *release.TagName

	if (release != nil) {
		fmt.Println("There is a new version available:", *release.TagName)
	} else {
		fmt.Println("There are no any new version available")
		return nil
	}

	fmt.Println("Current version is", ctx.Version)

	if !needUpgrade {
		fmt.Println("The application is up-to-date")
		return nil
	}

	if (args.Check) {
		return nil
	}
	
	if ctx.Version == "" || ctx.Version == "develop" && !args.Force {
		fmt.Println("Refusing to upgrade self-built application without --force")
		return nil
	}

	err = upgrade(client, *release.Assets[0].ID)
	if err != nil {
		return err
	}

	return nil
}

func upgrade(client *github.Client, id int64) error {

	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	execDir := filepath.Dir(execPath)

	dir := filepath.Join(execDir, ".self-upgrade")
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, "upgrade.zip")

	fmt.Println("Downloading ...")
	err = downloadLatestRelease(client, id, tmpfn)
	if err != nil {
		return err
	}

	fmt.Println("Extracting ...")

	r, err:= zip.OpenReader(tmpfn)	
	if err != nil {
		return err
	}

	defer r.Close()

	type Item struct {
		source string
		dest   string
		backup string
	}

	var items []Item

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return err
		}

		defer rc.Close()

		tmpfn := filepath.Join(dir, f.Name)
		outFile, err := os.Create(tmpfn)
		if err != nil {
			return err
		}

		defer outFile.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}

		outFile.Close()
		items = append(items, Item{
			source: tmpfn,
			dest:   filepath.Join(execDir, f.Name),
			backup: filepath.Join(execDir, f.Name) + ".old",
		})
	}

	backup := func() {
		for _, item := range items {
			os.Rename(item.dest, item.backup)
		}
	}

	apply := func() error {
		for _, item := range items {
			err := os.Rename(item.source, item.dest)
			if err != nil {
				return err
			}
		}
		return nil
	}

	rollback := func() {
		for _, item := range items {
			os.Rename(item.backup, item.dest)
		}
	}

	cleanup := func() {
		for _, item := range items {
			err = os.Remove(item.backup)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}

	defer cleanup()

	backup()
	err = apply()
	if err != nil {
		rollback()
		return err
	}

	return nil
}


func getLatestRelease(client *github.Client) (*github.RepositoryRelease, error) {

	release, resp, err := client.Repositories.GetLatestRelease(context2.Background(), "ProtocolONE", "cord.stool")
	if err != nil {
		return nil, fmt.Errorf("Repositories.GetLatestRelease returned error: %v\n%v", err, resp.Body)
	}

	if len(release.Assets) == 0 {
		return nil, nil
	}

	return release, nil
}

func downloadLatestRelease(client *github.Client, id int64, pathfile string) error {

	_, redir, err := client.Repositories.DownloadReleaseAsset(context2.Background(), "ProtocolONE", "cord.stool", id)
	if err != nil {
		return fmt.Errorf("Repositories.DownloadReleaseAsset returned error: %v", err)
	}

	resp2, err := http.Get(redir)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(resp2.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll returned error: %v", err)
	}

	if err := ioutil.WriteFile(pathfile, content, 0666); err != nil {
		return err
	}

	return nil
}

