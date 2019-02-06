package push

import (
	"fmt"

	"cord.stool/compressor/gzip"
	"cord.stool/context"
	"cord.stool/upload/akamai"
	"cord.stool/upload/cord"
	"cord.stool/upload/ftp"
	"cord.stool/upload/s3"
	"cord.stool/upload/sftp"

	"github.com/urfave/cli"
)

var args = struct {
	FtpUrl    string
	SftpUrl   string
	SourceDir string
	OutputDir string
	s3Args    s3.Args
	akmArgs   akamai.Args
	cordArgs  cord.Args
}{}

func Register(ctx *context.StoolContext) {

	cmd := cli.Command{
		Name:        "push",
		ShortName:   "p",
		Usage:       "Upload update",
		Description: "Upload update app bundle to one of servers",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "source, s",
				Usage:       "Path to game",
				Value:       "",
				Destination: &args.SourceDir,
			},
			cli.StringFlag{
				Name:        "output, o",
				Usage:       "Path to upload",
				Value:       "",
				Destination: &args.OutputDir,
			},
			cli.StringFlag{
				Name:        "ftp",
				Usage:       "Full ftp url path. Example ftp://user:password@host:port/upload/directory",
				Value:       "",
				Destination: &args.FtpUrl,
			},
			cli.StringFlag{
				Name:        "sftp",
				Usage:       "Full sftp url path. Example sftp://user:password@host:port/upload/directory",
				Value:       "",
				Destination: &args.SftpUrl,
			},
			cli.StringFlag{
				Name:        "aws-region",
				Usage:       "AWS region name",
				Value:       "",
				Destination: &args.s3Args.Region,
			},
			cli.StringFlag{
				Name:        "aws-credentials",
				Usage:       "Path to AWS credentials file",
				Value:       "",
				Destination: &args.s3Args.Credentials,
			},
			cli.StringFlag{
				Name:        "aws-profile",
				Usage:       "AWS profile name",
				Value:       "",
				Destination: &args.s3Args.Profile,
			},
			cli.StringFlag{
				Name:        "aws-id",
				Usage:       "AWS access key id",
				Value:       "",
				Destination: &args.s3Args.ID,
			},
			cli.StringFlag{
				Name:        "aws-key",
				Usage:       "AWS secret access key",
				Value:       "",
				Destination: &args.s3Args.Key,
			},
			cli.StringFlag{
				Name:        "aws-token",
				Usage:       "AWS session token",
				Value:       "",
				Destination: &args.s3Args.Token,
			},
			cli.StringFlag{
				Name:        "s3-bucket",
				Usage:       "Amazon S3 bucket name",
				Value:       "",
				Destination: &args.s3Args.S3Bucket,
			},
			cli.StringFlag{
				Name:        "akm-hostname",
				Usage:       "Akamai hostname",
				Value:       "",
				Destination: &args.akmArgs.Hostname,
			},
			cli.StringFlag{
				Name:        "akm-keyname",
				Usage:       "Akamai keyname",
				Value:       "",
				Destination: &args.akmArgs.Keyname,
			},
			cli.StringFlag{
				Name:        "akm-key",
				Usage:       "Akamai key",
				Value:       "",
				Destination: &args.akmArgs.Key,
			},
			cli.StringFlag{
				Name:        "akm-code",
				Usage:       "Akamai code",
				Value:       "",
				Destination: &args.akmArgs.Code,
			},
			cli.StringFlag{
				Name:        "cord-url",
				Usage:       "Cord Server url",
				Value:       "",
				Destination: &args.cordArgs.Url,
			},
			cli.StringFlag{
				Name:        "cord-login",
				Usage:       "Cord user login",
				Value:       "",
				Destination: &args.cordArgs.Login,
			},
			cli.StringFlag{
				Name:        "cord-password",
				Usage:       "Cord user password",
				Value:       "",
				Destination: &args.cordArgs.Password,
			},
			cli.BoolFlag{
				Name:        "cord-patch",
				Usage:       "Upload the difference between files",
				Destination: &args.cordArgs.Patch,
			},
			cli.BoolFlag{
				Name:        "cord-hash",
				Usage:       "Upload changed files only",
				Destination: &args.cordArgs.Hash,
			},
			cli.BoolFlag{
				Name:        "cord-wsync",
				Usage:       "Upload changed files only using Wharf protocol that enables incremental uploads",
				Destination: &args.cordArgs.Wharf,
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)

	gzip.Init()
}

func do(ctx *context.StoolContext, c *cli.Context) error {

	if args.SourceDir == "" {
		return fmt.Errorf("Path to game is required")
	}

	if args.FtpUrl == "" && args.SftpUrl == "" && args.s3Args.S3Bucket == "" && args.akmArgs.Hostname == "" && args.cordArgs.Url == "" {
		return fmt.Errorf("Specify one of following flags: ftp, sftp, s3-bucket, akm-hostname, cord-url")
	}

	if args.FtpUrl != "" {
		err := ftp.Upload(args.FtpUrl, args.SourceDir)
		if err != nil {
			return err
		}
	}

	if args.SftpUrl != "" {
		err := sftp.Upload(args.SftpUrl, args.SourceDir)
		if err != nil {
			return err
		}
	}

	if args.s3Args.S3Bucket != "" {
		args.s3Args.SourceDir = args.SourceDir
		args.s3Args.OutputDir = args.OutputDir
		err := s3.Upload(args.s3Args)
		if err != nil {
			return err
		}
	}

	if args.akmArgs.Hostname != "" {
		args.akmArgs.SourceDir = args.SourceDir
		args.akmArgs.OutputDir = args.OutputDir
		err := akamai.Upload(args.akmArgs)
		if err != nil {
			return err
		}
	}

	if args.cordArgs.Url != "" {
		args.cordArgs.SourceDir = args.SourceDir
		args.cordArgs.OutputDir = args.OutputDir
		err := cord.Upload(args.cordArgs)
		if err != nil {
			return err
		}
	}

	return nil
}
