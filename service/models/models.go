package models

type AppKey struct {
	PrivateKeyPath string `json:"private_key_path"`
	PublicKeyPath  string `json:"public_key_path"`
	JwtExpDelta    int    `json:"jwt_exp_delta"`
}

type DbAuth struct {
	Hosts    []string `json:"host"`
	Uname    string   `json:"username"`
	Pswd     string   `json:"password"`
	Database string   `json:"database"`
}

type Configuration struct {
	AppKeyCfg AppKey `json:"app_key"`
	DbAuthCfg DbAuth `json:"db_auth"`
}

type Authorisation struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
	Storage string `json:"storage" form:"username"`
}

type CreatedId struct {
	Title string `json:"title"`
	Id    string `json:"id"`
}

type AuthToken struct {
	ClientId string `json:"client_id"`
	Token    string `json:"token"`
}

type AuthRefresh struct {
	ClientId string `json:"client_id"`
}

type Error struct {
	Message string `json:"message"`
}

type CreateCmd struct {
	Source  string `json:"source"`
	Output  string `json:"output"`
	Archive bool   `json:archive`
}

type DiffCmd struct {
	SourceOld  string `json:"old"`
	SourceNew  string `json:"new"`
	OutputDiff string `json:"patch"`
}

type PushCmd struct {
	Source string `json:"source"`
	Output string `json:"output"`

	FtpUrl  string `json:"ftp"`
	SftpUrl string `json:"sftp"`

	AWSRegion      string `json:"aws-region"`
	AWSCredentials string `json:"aws-credentials"`
	AWSProfile     string `json:"aws-profile"`
	AWSID          string `json:"aws-id"`
	AWSKey         string `json:"aws-key"`
	AWSToken       string `json:"aws-token"`
	S3Bucket       string `json:"s3-bucket"`

	AkmHostname string `json:"akm-hostname"`
	AkmKeyname  string `json:"akm-keyname"`
	AkmKey      string `json:"akm-key"`
	AkmCode     string `json:"akm-code"`
}

type TorrentCmd struct {
	Source       string   `json:"source"`
	Target       string   `json:"target"`
	WebSeeds     []string `json:"web-seeds"`
	AnnounceList []string `json:"announce-list"`
	PieceLength  uint     `json:piece-length`
}
