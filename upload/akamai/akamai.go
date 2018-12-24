package akamai

import (
	"errors"
	"fmt"
	"net/textproto"
	"path/filepath"
	"strings"

	"cord.stool/utils"

	"github.com/gosuri/uiprogress"
	"github.com/akamai/netstoragekit-golang"
)

type Args = struct {
	SourceDir string
	OutputDir string
	Hostname  string
	Keyname   string
	Key       string
	Code      string
}

func Upload(args Args) error {

	fmt.Println("Uploading to Akamai CDN ...")
	uiprogress.Start()

	fullSourceDir, err := filepath.Abs(args.SourceDir)
	if err != nil {
		return err
	}

	fc, err := utils.FileCount(fullSourceDir)
	if err != nil {
		return err
	}

	uiprogress.Start()
	barTotal := uiprogress.AddBar(fc + 1 ).AppendCompleted().PrependElapsed()
	barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return "Total progress"
	})

	ns := netstorage.NewNetstorage(args.Hostname, args.Keyname, args.Key, false)

	rootPath := fmt.Sprintf("/%s/", args.Code)
	if args.OutputDir != "" {
		path := strings.Replace(args.OutputDir, "\\", "/", -1)
		err = mkdirRecursive(ns, rootPath, path)
		if err != nil {
			return err
		}
	}

	stopCh := make(chan struct{})
	defer func() {
		select {
		case stopCh <- struct{}{}:
		default:
		}

		close(stopCh)
	}()

	f, e := utils.EnumFilesRecursive(fullSourceDir, stopCh)

	bar := uiprogress.AddBar(3).AppendCompleted().PrependElapsed()

	var curTitle string
	var title *string
	title = &curTitle

	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return *title
	})

	barTotal.Incr();

	for path := range f {

		_, fn := filepath.Split(path)
		curTitle = fmt.Sprint("Uploading file: ", fn)

		barTotal.Incr();
		bar.Set(0);
		bar.Incr();

		relativePath, err := filepath.Rel(fullSourceDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(rootPath, args.OutputDir)
		destPath = filepath.Join(destPath, relativePath)
		destPath = strings.Replace(destPath, "\\", "/", -1)

		RootDir := filepath.Join(rootPath, args.OutputDir)
		RootDir = strings.Replace(RootDir, "\\", "/", -1)
		destDir, _ := filepath.Split(relativePath)
		destDir = strings.Replace(destDir, "\\", "/", -1)

		if destDir != "" {
			err := mkdirRecursive(ns, RootDir, destDir)
			if err != nil {
				return err
			}
		}

		bar.Incr();

		res, _, err := ns.Upload(path, destPath)
		if err != nil {
			return err
		}

		if res.StatusCode != 200 {
			return errors.New("Akamai Upload failed")
		}

		bar.Incr();
	}

	err = <-e
	if err != nil {
		return err
	}

	curTitle = "Finished"

	uiprogress.Stop()
	fmt.Println("Upload completed.")

	return nil
}

func mkdirRecursive(ns *netstorage.Netstorage, root string, dir string) error {
	dir = filepath.ToSlash(dir)
	dirs := strings.Split(dir, "/")

	tmpDir := ""

	for _, d := range dirs {
		if d == "" {
			continue
		}

		tmpDir += d + "/"

		path := filepath.Join(root , tmpDir)
		path = strings.Replace(path, "\\", "/", -1)

		res, _, err := ns.Mkdir(path)
		if err != nil {
			code := err.(*textproto.Error).Code
			if code == 550 {
				continue
			}
			return err
		}
		if res.StatusCode != 200 {
			return errors.New("Akamai Mkdir failed")
		}
	}

	return nil
}
