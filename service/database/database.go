package database

import (
    "cord.stool/service/config"
    "cord.stool/service/models"

    mgo "gopkg.in/mgo.v2"
    "go.uber.org/zap"
    "gopkg.in/mgo.v2/bson"
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

    dbConf.Dbs = session
    zap.S().Infof("Connected to DB: \"%s\" [u:\"%s\":p\"%s\"]", dbConf.Database, dbConf.Username, dbConf.Password)
    return nil
}

type DBManager interface {
	FindByName(string) (*models.User, error)
    RemoveByName(string) error
    Insert(*models.User) error
}

type UserManager struct {
	collection *mgo.Collection
}

func NewUserManager() (*UserManager) {
    session := dbConf.Dbs.Copy()
    return &UserManager{collection: session.DB(dbConf.Database).C("users")}
}

func (manager *UserManager) FindByName(name string) ([]*models.User, error) {

    var dbUsers []*models.User
    err := manager.collection.Find(bson.M{"username": name}).All(&dbUsers)
    if err != nil {
		return nil, err
	}

	return dbUsers, nil
}

func (manager *UserManager) RemoveByName(name string) error {

    err := manager.collection.Remove(bson.M{"username": name})
    if err != nil {
		return err
	}

	return nil
}

func (manager *UserManager) Insert(user *models.User) error {

    err := manager.collection.Insert(user)
    if err != nil {
		return err
	}

	return nil
}
