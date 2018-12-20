package push

import (
	"fmt"

	"cord.stool/context"
	"cord.stool/upload/ftp"
	"cord.stool/upload/sftp"
	"cord.stool/upload/s3"
	"cord.stool/upload/akamai"

	"github.com/urfave/cli"
)

var args = struct {
	FtpUrl    string
	SftpUrl    string
	SourceDir string
	OutputDir string
	s3Args s3.Args
	akmArgs akamai.Args
}{}

func Register(ctx *context.StoolContext) {
	cmd := cli.Command{
		Name:        	"push",
		ShortName:		"p",
		Usage:			"Upload update",
		Description:	"Upload update app bundle to one of servers",

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
				Destination: &args.s3Args.AWSRegion,
			},
			cli.StringFlag{
				Name:        "aws-credentials",
				Usage:       "Path to AWS credentials file",
				Value:       "",
				Destination: &args.s3Args.AWSCredentials,
			},
			cli.StringFlag{
				Name:        "aws-profile",
				Usage:       "AWS profile name",
				Value:       "",
				Destination: &args.s3Args.AWSProfile,
			},
			cli.StringFlag{
				Name:        "aws-id",
				Usage:       "AWS access key id",
				Value:       "",
				Destination: &args.s3Args.AWSID,
			},
			cli.StringFlag{
				Name:        "aws-key",
				Usage:       "AWS secret access key",
				Value:       "",
				Destination: &args.s3Args.AWSKey,
			},
			cli.StringFlag{
				Name:        "aws-token",
				Usage:       "AWS session token",
				Value:       "",
				Destination: &args.s3Args.AWSToken,
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
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func do(ctx *context.StoolContext, c *cli.Context) error {
	
	if args.SourceDir == "" {
		return fmt.Errorf("Path to game is required")
	}

	if args.FtpUrl == "" && args.SftpUrl == "" && args.s3Args.S3Bucket == "" && args.akmArgs.Hostname == "" {
		return fmt.Errorf("Specify one of following flags: ftp, sftp, s3-bucket, akm-hostname")
	}
	
	if args.FtpUrl != "" {
		err := ftp.UploadToFTP(args.FtpUrl, args.SourceDir)
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

	return nil
}
