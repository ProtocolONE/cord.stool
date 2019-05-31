package models

import (
	"time"
)

const (
	ErrorInvalidJSONFormat         = 1
	ErrorDatabaseFailure           = 2
	ErrorAlreadyExists             = 3
	ErrorGenUserStorageName        = 4
	ErrorInvalidUsernameOrPassword = 5
	ErrorGenToken                  = 6
	ErrorLogout                    = 7
	ErrorGetUserStorage            = 8
	ErrorFileIOFailure             = 9
	ErrorApplyPatch                = 10
	ErrorUnauthorized              = 11
	ErrorTokenExpired              = 12
	ErrorInvalidToken              = 13
	ErrorLoginTracker              = 14
	ErrorAddTorrent                = 15
	ErrorDeleteTorrent             = 16
	ErrorWharfLibrary              = 17
	ErrorInvalidRequest            = 18
	ErrorNotFound                  = 19
	ErrorInternalError             = 20
	ErrorCreateTorrent             = 21
	ErrorBuildIsNotPublished       = 22
	ErrorInvalidPlatformName       = 23
	ErrorInvalidBuildPlatform      = 24
)

const (
	Win64Mask    = 0
	Win32Mask    = 1
	Win32_64Mask = 2
	MacOSMask    = 4
	LinuxMask    = 8
)

const (
	Win64    = "win64"
	Win32    = "win32"
	Win32_64 = "win32_64"
	MacOS    = "macos"
	Linux    = "linux"
)

const (
	Win64_Folder    = ".win64"
	Win32_Folder    = ".win32"
	Win32_64_Folder = ".win32_64"
	MacOS_Folder    = ".macos"
	Linux_Folder    = ".linux"
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
	BuildID  string `json:"build_id"`
	FilePath string `json:"filepath"`
	FileName string `json:"filename"`
	FileData []byte `json:"filedata"`
	Patch    bool   `json:"patch"`
	Config   bool   `json:"config"`
	Platform string `json:"platform"`
}

type CompareHashCmd struct {
	BuildID  string `json:"build_id"`
	FilePath string `json:"filepath"`
	FileName string `json:"filename"`
	FileHash string `json:"filehash"`
	Platform string `json:"platform"`
}

type CompareHashCmdResult struct {
	Equal bool `json:"equal"`
}

type SignatureCmd struct {
	BuildID string `json:"build_id"`
}

type SignatureCmdResult struct {
	FileData []byte `json:"filedata"`
}

type ApplyPatchCmd struct {
	BuildID    string `json:"build_id"`
	SrcBuildID string `json:"src_build_id"`
	FileData   []byte `json:"filedata"`
	Platform   string `json:"platform"`
}

type User struct {
	Username string `bson:"username" json:"username"`
	Password string `bson:"password" json:"password"`
	Storage  string `bson:"storage" json:"storage"`
}

type TorrentCmd struct {
	InfoHash string `bson:"info_hash" json:"info_hash"`
}

type Branch struct {
	ID        string    `bson:"_id" json:id`
	Name      string    `json:"name"`
	GameID    string    `json:"game_id"`
	LiveBuild string    `json:"live_build"`
	Live      bool      `json:"live"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

type Build struct {
	ID       string    `bson:"_id" json:id`
	BranchID string    `json:"branch_id"`
	Created  time.Time `json:"created"`
	//Platform string    `json:"platform"`
}

type Depot struct {
	ID       string    `bson:"_id" json:id`
	Created  time.Time `json:"created"`
	Platform string    `json:"platform"`
}

type BuildDepot struct {
	ID       string    `bson:"_id" json:id`
	BuildID  string    `json:"build_id"`
	DepotID  string    `json:"depot_id"`
	LinkID   string    `json:"link_id"`
	Platform string    `json:"platform"`
	Created  time.Time `json:"created"`
}

type ShallowBranchCmdResult struct {
	SourceID   string `json:"source_id"`
	SourceName string `json:"source_name"`
	TargetID   string `json:"target_id"`
	TargetName string `json:"target_name"`
}

type GameGenre struct {
	Main     int64   `json:"main"`
	Addition []int64 `json:"addition" validate:"required"`
}

type GamePrice struct {
	Price    float64 `json:"price" validate:"required"`
	Currency string  `json:"currency" validate:"required"`
}

type Game struct {
	ID           string    `json:"id"`
	InternalName string    `json:"internalName"`
	Icon         string    `json:"icon"`
	Genres       GameGenre `json:"genres"`
	ReleaseDate  time.Time `json:"releaseDate"`
	Prices       GamePrice `json:"prices"`
}

type Vendor struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Domain3         string `json:"domain3"`
	Email           string `json:"email"`
	ManagerID       string `json:"manager_id"`
	HowManyProducts string `json:"howmanyproducts"`
}

type GameInfo struct {
	ID                   string    `json:"id"`
	InternalName         string    `json:"internalName"`
	Title                string    `json:"title"`
	Developers           string    `json:"developers"`
	Publishers           string    `json:"publishers"`
	ReleaseDate          time.Time `json:"releaseDate" validate:"required"`
	DisplayRemainingTime bool      `json:"displayRemainingTime"`
	AchievementOnProd    bool      `json:"achievementOnProd"`
}

type UpdateInfo struct {
	BuildID string   `json:"build_id"`
	Config  string   `json:"config"`
	Files   []string `json:"files"`
}

type DownloadCmd struct {
	FilePath string `json:"filepath"`
	FileData []byte `json:"filedata"`
}

type ConfigLocales struct {
	Label     string `json:"label"`
	Locale    string `json:"locale"`
	LocalRoot string `json:"local_root"`
}

type ConfigManifest struct {
	Label     string          `json:"label"`
	Platform  string          `json:"platform"`
	Locales   []ConfigLocales `json:"locales"`
	LocalRoot string          `json:"local_root"`
}

type ConfigApplication struct {
	ID        float64          `json:"id"`
	Manifests []ConfigManifest `json:"manifests"`
}

type Config struct {
	Application ConfigApplication `json:"application"`
}
