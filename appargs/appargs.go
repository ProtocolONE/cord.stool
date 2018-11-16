package appargs

import (
	"flag"
	"os"
)

var (
	// SourceDir ...
	SourceDir string
	// OutputDir ...
	OutputDir string
	// OutputFile ...
	OutputFile string
	// FTPUrl ...
	FTPUrl string
	// AWSRegion ...
	AWSRegion string
	// AWSCredentials ...
	AWSCredentials string
	// AWSProfile ...
	AWSProfile string
	// AWSID ...
	AWSID string
	// AWSKey ...
	AWSKey string
	// S3BucketName ...
	S3BucketName string
	// CheckExtension ...
	CheckExtension string
	// ImportantFiles ...
	ImportantFiles string
	// TorrentConfig ...
	TorrentConfig string
	// RegistryConfig ...
	RegistryConfig string
	// IgnoreFiles ...
	IgnoreFiles string
	// NoArchive ...
	NoArchive bool
	// IgnoreHidden ...
	IgnoreHidden string
	// Threads ...
	Threads int
	// FingerprintMd5 ...
	FingerprintMd5 bool
	// VersionFilePath ...
	VersionFilePath string
	// VersionFilePath1 ...
	VersionFilePath1 string
	// VersionFilePath2 ...
	VersionFilePath2 string
)

// Init ...
func Init() bool {

	flag.StringVar(&SourceDir, "sourceDir", "", "Source dirrectory path.")
	flag.StringVar(&OutputDir, "outputDir", "", "Output dirrectory path.")
	flag.StringVar(&OutputFile, "outputFile", "update.crc", "Output crc file name. Default value: 'update.crc'.")
	flag.StringVar(&FTPUrl, "ftpUrl", "", "Upload files to FTP. Specify ftp url, example: ftp://login:password@ftp.myserver.com:9090/path.")
	flag.StringVar(&AWSRegion, "awsRegion", "", "AWS region name.")
	flag.StringVar(&AWSCredentials, "awsCredentials", "", "Path to AWS credentials file.")
	flag.StringVar(&AWSProfile, "awsProfile", "", "AWS profile name.")
	flag.StringVar(&AWSID, "awsId", "", "AWS access key id.")
	flag.StringVar(&AWSKey, "awsKey", "", "AWS secret access key.")
	flag.StringVar(&S3BucketName, "s3BucketName", "", "Amazon S3 bucket name.")
	flag.StringVar(&CheckExtension, "checkExtension", "exe,dll,bin", "Comma separated file extension for important file. Default value: 'exe,dll,bin'.")
	flag.StringVar(&ImportantFiles, "importantFiles", "", "Comma separated file list for mark as important file.")
	flag.StringVar(&TorrentConfig, "torrentConfig", "", "Torrent config file.")
	flag.StringVar(&RegistryConfig, "registryConfig", "update.reg", "Registry file to import. Default value: 'update.reg'.")
	flag.StringVar(&IgnoreFiles, "ignoreFiles", "", "Ignore file list with comma as delimiter.")
	flag.BoolVar(&NoArchive, "noArchive", false, "Don't archive files.")
	flag.StringVar(&IgnoreHidden, "ignoreHidden", "", "Skip hiden files in update.crc and torrent.")
	flag.IntVar(&Threads, "threads", 1, "Number of thread during update.crc hashing. Default value: '1'.")
	flag.BoolVar(&FingerprintMd5, "fingerprintMd5", false, "If set update will contain <filename>.md5 finger print for each file  Default value: 'false'.")
	flag.StringVar(&VersionFilePath, "versionFilePath", "", "Relative path to file PE with VersionInfo.")
	flag.StringVar(&VersionFilePath1, "versionFilePath1", "", "Relative path to file PE with VersionInfo.")
	flag.StringVar(&VersionFilePath2, "versionFilePath2", "", "Relative path to file PE with VersionInfo.")

	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		return false
	}

	return true
}
