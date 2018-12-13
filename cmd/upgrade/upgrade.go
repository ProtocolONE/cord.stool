package upgrade

import (
	"fmt"
	"strings"
	"strconv"
	"path/filepath"
	"os"
	"time"
	"io"
	"io/ioutil"
	"net/http"
	"archive/zip"
	context2 "context"
	"cord.stool/context"
	"cord.stool/utils"
	"github.com/urfave/cli"
	"github.com/google/go-github/github"
)

var args = struct {
	force bool
	check bool
	list bool
	version string
	fileList cli.StringSlice
}{
	force: false,
	check: false,
	list: false,
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
				Destination: &args.force,
			},
			cli.BoolFlag{
				Name:        "check, c",
				Usage:       "Checking for new version",
				Destination: &args.check,
			},
			cli.BoolFlag{
				Name:        "list, l",
				Usage:       "Show all new available versions",
				Destination: &args.list,
			},
			cli.StringFlag{
				Name:        "ver, v",
				Usage:       "Upgrades application to this version",
				Value:       "",
				Destination: &args.version,
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)

	cmd = cli.Command{
		Name:       "upgrade_complete",
		HideHelp:   true,
		Hidden:     true,

		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "file, f",
				Value: &args.fileList,
				Hidden: true,
			},
		},
		Action: func(c *cli.Context) error {
			return doComplete(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func do(ctx *context.StoolContext, c *cli.Context) error {

	client := github.NewClient(nil)

	if args.list {
		releases, err := getReleaseList(client)
		if err != nil {
			return err
		}

		if releases == nil || len(releases) == 0 {
			fmt.Println("There are no any new version available")
			return nil
		}

		var versions []string

		for _, release := range releases {
			if !*release.Prerelease {
				versions = append(versions, *release.TagName)
			}
		}

		if len(versions) == 1 {
			fmt.Println("Found one new version: ")
		} else if len(versions) > 1 {
			fmt.Printf("Found %d new versions:\n", len(releases))
		} else {
			fmt.Println("There are no any new version available")
		}

		for _, versions := range versions {
			fmt.Println(versions)
		}
		return nil
	}

	var release *github.RepositoryRelease
	var err error

	if args.version != "" {

		if compareVersion(ctx.Version, args.version) < 0 {
			fmt.Printf("Current version %s is newer version than %s\n", ctx.Version, args.version)
			return nil
		}

		release, err = getRelease(client, args.version)
		if err != nil {
			return err
		}

		if release != nil && compareVersion(args.version, *release.TagName) == 0 {
			fmt.Printf("Found version %s\n", *release.TagName)
		} else {
			fmt.Printf("The version %s is not found\n", args.version)
			return nil
		}
		//release2 = *release

	} else {

		release, err = getLatestRelease(client)
		if err != nil {
			return err
		}

		if (release != nil) {
			fmt.Println("There is a new version available:", *release.TagName)
		} else {
			fmt.Println("There are no any new version available")
			return nil
		}
		//release2 = *release
	}

	fmt.Println("Current version is", ctx.Version)

	needUpgrade := compareVersion(ctx.Version, *release.TagName) > 0
	if !needUpgrade {
		fmt.Println("The application is up-to-date")
		return nil
	}

	if (args.check) {
		return nil
	}
	
	if ctx.Version == "" || ctx.Version == "develop" && !args.force {
		fmt.Println("Refusing to upgrade self-built application without --force")
		return nil
	}

	fmt.Printf("Upgrading application to %s version\n", *release.TagName)
	err = upgrade(client, *release.Assets[0].ID)
	if err != nil {
		return err
	}
	
	fmt.Printf("Application is upgraded from %s to %s\n", ctx.Version, *release.TagName)

	return nil
}

func upgrade(client *github.Client, id int64) error {

	tmpDir, err := ioutil.TempDir(os.TempDir(), "p1-")	
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpfn := filepath.Join(tmpDir, "upgrade.zip")

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
		dest   string
		backup string
	}
	
	var backups []string
	complete := func() {
		utils.CompleteUpgrade(backups)
	}
	defer complete()

	var items []Item
	rollback := func() {
		for _, item := range items {
			os.Rename(item.backup, item.dest)
		}
	}
	defer rollback()

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(execPath)

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		dest := filepath.Join(execDir, f.Name)
		backup := dest + ".old"

		os.Rename(dest, backup)
		if err != nil {
			return err
		}
		backups = append(backups, backup)

		outFile, err := os.Create(dest)
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
			dest:   dest,
			backup: backup,
		})
	}

	items = nil
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

func compareVersion(ver1 string, ver2 string) int {
	
	v1 := strings.Split(ver1, ".")
	v2 := strings.Split(ver2, ".")

	v10, v11, v12, v20, v21, v22 := 0, 0, 0, 0, 0, 0

	if len(v1) >= 1 {
		v10, _ = strconv.Atoi(v1[0])
		if len(v1) >= 2 {
			v11, _ = strconv.Atoi(v1[1])
			if len(v1) >= 3 {
				v12, _ = strconv.Atoi(v1[2])
			}
		}
	}

	if len(v2) >= 1 {
		v20, _ = strconv.Atoi(v2[0])
		if len(v2) >= 2 {
			v21, _ = strconv.Atoi(v2[1])
			if len(v2) >= 3 {
				v22, _ = strconv.Atoi(v2[2])
			}
		}
	}

	if v10 > v20 {
		return -1
	} else if v10 < v20 {
		return 1
	} else if v11 > v21 {
		return -1
	} else if v11 < v21 {
		return 1
	} else if v12 > v22 {
		return -1
	} else if v12 < v22 {
		return 1
	} else {
		return 0
	}
}

func getRelease(client *github.Client, ver string) (*github.RepositoryRelease, error) {

	releases, err := getReleaseList(client)
	if err != nil {
		return nil, err
	}

	if releases == nil || len(releases) == 0 {
		return nil, nil
	}

	for _, release := range releases {
		if ver == *release.TagName && len(release.Assets) > 0 && !*release.Prerelease { 
			return release, nil
		}
	}

	return nil, nil
}

func getReleaseList(client *github.Client) ([]*github.RepositoryRelease, error) {

	opt := &github.ListOptions{Page: 1, PerPage: 20}
	releases, resp, err := client.Repositories.ListReleases(context2.Background(), "ProtocolONE", "cord.stool", opt)
	if err != nil {
		return nil, fmt.Errorf("Repositories.ListReleases returned error: %v\n%v", err, resp.Body)
	}

	return releases, nil
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

func doComplete(ctx *context.StoolContext, c *cli.Context) error {
	
	time.Sleep(1 * time.Second)
	
	for _, f := range args.fileList {
		os.Remove(f)
	}
	return nil
}