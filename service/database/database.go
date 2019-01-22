package database

import (
    "cord.stool/service/config"
    //migrate "github.com/xakep666/mongo-migrate"
    //_ "management.api/migrations"
    mgo "gopkg.in/mgo.v2"
    //"gopkg.in/mgo.v2/bson"
    "go.uber.org/zap"
)

type DbConf struct {
    Dbs      *mgo.Session
    Addrs    []string
    Database string
    Username string
    Password string
}

var dbConf *DbConf

func Init() (error) {
    cfg := config.Get().Database
    dbConf = &DbConf{
        Addrs:    []string{cfg.Host},
        Database: cfg.Database,
        Username: cfg.User,
        Password: cfg.Password,
    }
    session, err := mgo.DialWithInfo(&mgo.DialInfo{
        Addrs:    dbConf.Addrs,
        Database: dbConf.Database,
        Username: dbConf.Username,
        Password: dbConf.Password,
    })
    if err != nil {
        zap.S().Fatal(err)
        return err
    }
    //db := session.DB("")
    //m := migrate.NewMigrate(db, migrate.Migration{
    //    Version: 1,
    //    Description: "Add user index",
    //    Up: func(db *mgo.Database) error {
    //        return db.C("users").EnsureIndex(mgo.Index{Name: "username", Key: []string{"username"}})
    //    },
    //    Down: func(db *mgo.Database) error {
    //        return db.C("users").DropIndexName("username")
    //    },
    //})
    //if err := m.Up(migrate.AllAvailable); err != nil {
    //    zap.S().Fatal(err)
    //    return err
    //}
    dbConf.Dbs = session
    zap.S().Infof("Connected to DB: \"%s\" [u:\"%s\":p\"%s\"]", dbConf.Database, dbConf.Username, dbConf.Password)
    return nil
}

func Get(collection string) (*mgo.Collection) {
    session := dbConf.Dbs.Copy()
    //defer session.Close()
    return session.DB(dbConf.Database).C(collection)
}
