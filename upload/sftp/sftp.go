package sftp

import (
	"net/url"
	"path/filepath"
	"strings"
	"fmt"
	"os"

	"cord.stool/utils"

	"github.com/gosuri/uiprogress"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Upload ...
func Upload(sftpUrl, sourceDir string) error {

	fmt.Println("Uploading to SFTP Server ...")

	fullSourceDir, err := filepath.Abs(sourceDir)
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

	u, err := url.Parse(sftpUrl)
	if err != nil {
		return errors.Wrap(err, "Invalid sftp url")
	}

	if !strings.EqualFold(u.Scheme, "sftp") {
		return errors.New("UploadToFTP: Invalid url scheme ")
	}

	var login = ""
	var password = ""
	if u.User != nil {
		login = u.User.Username()
		password, _ = u.User.Password()
	}

	var auths []ssh.AuthMethod

	if password != "" {
		auths = append(auths, ssh.Password(password))
	}

	config := ssh.ClientConfig{
		User: login,
		Auth: auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", u.Host, &config)
	if err != nil {
		fmt.Println(err.Error())
		return errors.Wrap(err, "Failed to coonect")
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn, sftp.MaxPacket(1<<15))
	if err != nil {
		return errors.Wrap(err, "Failed to create client")
	}
	defer client.Close()

	sftpRoot := u.Path
	if sftpRoot[0] == '\\' || sftpRoot[0] == '/' {
		sftpRoot = sftpRoot[1:]
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

	bar := uiprogress.AddBar(5).AppendCompleted().PrependElapsed()

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

		file, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "Cannot open file:")
		}
		defer file.Close()
		
		bar.Incr();

		destPath := filepath.Join(sftpRoot, relativePath)
		destPath = strings.Replace(destPath, "\\", "/", -1)

		destDir, _ := filepath.Split(destPath)
		err =  client.MkdirAll(destDir)
		if err != nil {
			return errors.Wrap(err, "Cannot create directory:")
		}

		bar.Incr();

		w, err := client.Create(destPath)
		if err != nil {
			return errors.Wrap(err, "Cannot create file:")
		}
		defer w.Close()

		bar.Incr();

		_, err = w.WriteTo(file)
		if err != nil {
			return errors.Wrap(err, "Write file failed:")
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
