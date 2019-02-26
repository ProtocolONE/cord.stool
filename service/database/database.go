package database

import (
	"cord.stool/service/config"
	"cord.stool/service/models"

	"go.uber.org/zap"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type DbConf struct {
	Dbs      *mgo.Session
	Database string
}

var dbConf *DbConf

func Init() error {

	cfg := config.Get().Database

	dbConf = &DbConf{
		Database: cfg.Database,
	}

	session, err := mgo.Dial(cfg.Host)
	if err != nil {
		zap.S().Fatal(err)
		return err
	}

	db := session.DB(cfg.Database)
	err = db.Login(cfg.User, cfg.Password)
	if err != nil {
		zap.S().Fatal(err)
		return err
	}

	dbConf.Dbs = session
	zap.S().Infof("Connected to DB: \"%s\" [u:\"%s\":p\"%s\"]", dbConf.Database, cfg.User, cfg.Password)

	return nil
}

type UserManager struct {
	collection *mgo.Collection
}

func NewUserManager() *UserManager {
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
