package models

import (
	"time"
)

const (
	ErrorInvalidJSONFormat         = 1
	ErrorReadDataBase              = 2
	ErrorUserAlreadyExists         = 3
	ErrorGenUserStorageName        = 4
	ErrorCreateUser                = 5
	ErrorDeleteUser                = 6
	ErrorInvalidUsernameOrPassword = 7
	ErrorGenToken                  = 8
	ErrorLogout                    = 9
	ErrorGetUserStorage            = 10
	ErrorCreatePath                = 11
	ErrorGenTempFile               = 12
	ErrorWriteFile                 = 13
	ErrorApplyPatch                = 14
	ErrorUnauthorized              = 15
	ErrorTokenExpired              = 16
	ErrorInvalidToken              = 17
	ErrorLoginTracker              = 18
	ErrorAddTracker                = 19
	ErrorDeleteTracker             = 20
	ErrorWharfLibrary              = 21
	ErrorCreateFile                = 22
	ErrorReadFile                  = 23
	ErrorOpenFile                  = 24
	ErrorInvalidRequest            = 25
)

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

type Authorization struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type CreatedId struct {
	Title string `json:"title"`
	Id    string `json:"id"`
}

type AuthToken struct {
	ClientId     string `json:"client_id"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthRefresh struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type UploadCmd struct {
	FilePath string `json:"filepath"`
	FileName string `json:"filename"`
	FileData []byte `json:filedata`
	Patch    bool   `json:patch`
}

type CompareHashCmd struct {
	FilePath string `json:"filepath"`
	FileName string `json:"filename"`
	FileHash string `json:filehash`
}

type CompareHashCmdResult struct {
	Equal bool `json:equal`
}

type SignatureCmd struct {
	Path string `json:"path"`
}

type SignatureCmdResult struct {
	FileData []byte `json:filedata`
}

type ApplyPatchCmd struct {
	Path     string `json:"path"`
	FileData []byte `json:filedata`
}

type User struct {
	Username string `bson:"username" json:"username"`
	Password string `bson:"password" json:"password"`
	Storage  string `bson:"storage" json:"storage"`
}

type TorrentCmd struct {
	InfoHash string `bson:"info_hash" json:"info_hash"`
}

type BranchInfoCmd struct {
	ID     string `json:id`
	Name   string `json:name`
	GameID string `json:"game_id"`
}

type ListBranchCmd struct {
	GameID string `json:"game_id"`
}

type Branch struct {
	ID          string    `json:id`
	Name        string    `json:name`
	GameID      string    `json:"game_id"`
	LiveBuildID string    `json:live_build_id`
	Created     time.Time `json:created`
}

type ListBranchCmdResult struct {
	List []Branch `json:"list"`
}

type ShallowBranchCmd struct {
	SourceNameOrID string `json:source_name_or_id`
	TargetNameOrID string `json:target_name_or_id`
}

type ShallowBranchCmdResult struct {
	SourceID   string `json:source_id`
	SourceName string `json:source_name`
	TargetID   string `json:target_id`
	TargetName string `json:target_name`
}
