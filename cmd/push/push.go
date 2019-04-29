package push

import (
	"fmt"
	"net/url"
	"strings"

	"cord.stool/context"
	"cord.stool/upload/akamai"
	"cord.stool/upload/ftp"
	"cord.stool/upload/s3"
	"cord.stool/upload/sftp"

	"github.com/urfave/cli"
)

var args = struct {
	FtpUrl    string
	SourceDir string
	OutputDir string
	s3Args    s3.Args
	akmArgs   akamai.Args
}{}

func Register(ctx *context.StoolContext) {

	cmd := cli.Command{
		Name:        "push",
		ShortName:   "p",
		Usage:       "Upload files",
		Description: "Upload app files to one of servers",

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
		},

		Subcommands: cli.Commands{
			cli.Command{
				Name:        "ftp",
				Usage:       "Upload to FTP server",
				Description: "Upload to FTP server",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "url",
						Usage:       "Full FTP or SFTP url. Example ftp://user:password@host:port/upload/directory",
						Value:       "",
						Destination: &args.FtpUrl,
					},
				},
				Action: func(c *cli.Context) error {
					return doFtp(ctx, c)
				},
			},

			cli.Command{
				Name:        "aws",
				Usage:       "Upload to Amazon S3",
				Description: "Upload to Amazon S3",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "region",
						Usage:       "AWS region name",
						Value:       "",
						Destination: &args.s3Args.Region,
					},
					cli.StringFlag{
						Name:        "credentials",
						Usage:       "Path to AWS credentials file",
						Value:       "",
						Destination: &args.s3Args.Credentials,
					},
					cli.StringFlag{
						Name:        "profile",
						Usage:       "AWS profile name",
						Value:       "",
						Destination: &args.s3Args.Profile,
					},
					cli.StringFlag{
						Name:        "id",
						Usage:       "AWS access key id",
						Value:       "",
						Destination: &args.s3Args.ID,
					},
					cli.StringFlag{
						Name:        "key",
						Usage:       "AWS secret access key",
						Value:       "",
						Destination: &args.s3Args.Key,
					},
					cli.StringFlag{
						Name:        "token",
						Usage:       "AWS session token",
						Value:       "",
						Destination: &args.s3Args.Token,
					},
					cli.StringFlag{
						Name:        "bucket",
						Usage:       "Amazon S3 bucket name",
						Value:       "",
						Destination: &args.s3Args.S3Bucket,
					},
				},
				Action: func(c *cli.Context) error {
					return doAws(ctx, c)
				},
			},

			cli.Command{
				Name:        "akm",
				Usage:       "Upload to Akamai server",
				Description: "Upload to Akamai server",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "host",
						Usage:       "Akamai hostname",
						Value:       "",
						Destination: &args.akmArgs.Hostname,
					},
					cli.StringFlag{
						Name:        "name",
						Usage:       "Akamai keyname",
						Value:       "",
						Destination: &args.akmArgs.Keyname,
					},
					cli.StringFlag{
						Name:        "key",
						Usage:       "Akamai key",
						Value:       "",
						Destination: &args.akmArgs.Key,
					},
					cli.StringFlag{
						Name:        "code",
						Usage:       "Akamai code",
						Value:       "",
						Destination: &args.akmArgs.Code,
					},
				},
				Action: func(c *cli.Context) error {
					return doAkm(ctx, c)
				},
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func doFtp(ctx *context.StoolContext, c *cli.Context) error {

	if args.SourceDir == "" {
		return fmt.Errorf("-source flag is required")
	}

	if args.FtpUrl != "" {
		u, err := url.Parse(args.FtpUrl)
		if err != nil {
			return fmt.Errorf("Invalid ftp url")
		}

		if !strings.EqualFold(u.Scheme, "ftp") {

			err = ftp.Upload(args.FtpUrl, args.SourceDir)
			if err != nil {
				return err
			}

		} else if !strings.EqualFold(u.Scheme, "sftp") {

			err := sftp.Upload(args.FtpUrl, args.SourceDir)
			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("-url flag is required")
	}

	return nil
}

func doAws(ctx *context.StoolContext, c *cli.Context) error {

	if args.SourceDir == "" {
		return fmt.Errorf("-source flag is required")
	}

	if args.s3Args.S3Bucket == "" {
		return fmt.Errorf("-bucket flag is required")
	}

	args.s3Args.SourceDir = args.SourceDir
	args.s3Args.OutputDir = args.OutputDir
	err := s3.Upload(args.s3Args)
	if err != nil {
		return err
	}

	return nil
}

func doAkm(ctx *context.StoolContext, c *cli.Context) error {

	if args.SourceDir == "" {
		return fmt.Errorf("-source flag is required")
	}

	if args.akmArgs.Hostname == "" {
		return fmt.Errorf("-host flag is required")
	}

	args.akmArgs.SourceDir = args.SourceDir
	args.akmArgs.OutputDir = args.OutputDir
	err := akamai.Upload(args.akmArgs)
	if err != nil {
		return err
	}

	return nil
}

func do(ctx *context.StoolContext, c *cli.Context) error {

	return fmt.Errorf("Specify one of following sub-commands: ftp, aws or akm")
}
