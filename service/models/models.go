package models

type AppKey struct {
    PrivateKeyPath string `json:"private_key_path"`
    PublicKeyPath  string `json:"public_key_path"`
    JwtExpDelta    int    `json:"jwt_exp_delta"`
}

type DbAuth struct {
    Hosts      []string `json:"host"`
    Uname      string   `json:"username"`
    Pswd       string   `json:"password"`
    Database   string   `json:"database"`
}

type Configuration struct {
    AppKeyCfg AppKey `json:"app_key"`
    DbAuthCfg DbAuth `json:"db_auth"`
}

type Authorisation struct {
    Username string `json:"username" form:"username"`
    Password string `json:"password" form:"password"`
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
