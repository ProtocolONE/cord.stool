package database

import (
	"cord.stool/service/config"
	"cord.stool/service/models"

	"fmt"
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

	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{cfg.Host},
		Database: cfg.Database,
		Username: cfg.User,
		Password: cfg.Password,
	})
	if err != nil {
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

type BranchManager struct {
	collection *mgo.Collection
}

func NewBranchManager() *BranchManager {
	session := dbConf.Dbs.Copy()
	return &BranchManager{collection: session.DB(dbConf.Database).C("branches")}
}

func (manager *BranchManager) FindByID(id string) (*models.Branch, error) {

	var dbBranch []*models.Branch
	err := manager.collection.Find(bson.M{"_id": id}).All(&dbBranch)
	if err != nil {
		return nil, err
	}

	if dbBranch == nil {
		return nil, nil
	}

	if len(dbBranch) > 1 {
		return nil, fmt.Errorf("Database integrity error")
	}

	return dbBranch[0], nil
}

func (manager *BranchManager) FindByName(name string, gid string) (*models.Branch, error) {

	var dbBranch []*models.Branch
	err := manager.collection.Find(bson.M{"name": name, "gameid": gid}).All(&dbBranch)
	if err != nil {
		return nil, err
	}

	if dbBranch == nil {
		return nil, nil
	}

	if len(dbBranch) > 1 {
		return nil, fmt.Errorf("Database integrity error")
	}

	return dbBranch[0], nil
}

func (manager *BranchManager) Insert(branch *models.Branch) error {

	err := manager.collection.Insert(branch)
	if err != nil {
		return err
	}

	return nil
}

func (manager *BranchManager) Update(branch *models.Branch) error {

	err := manager.collection.UpdateId(branch.ID, branch)
	if err != nil {
		return err
	}

	return nil
}

func (manager *BranchManager) RemoveByID(id string) error {

	err := manager.collection.Remove(bson.M{"_id": id})
	if err != nil {
		return err
	}

	return nil
}

func (manager *BranchManager) List(gameID string) ([]*models.Branch, error) {

	var dbBranch []*models.Branch
	err := manager.collection.Find(bson.M{"gameid": gameID}).All(&dbBranch)
	if err != nil {
		return nil, err
	}

	return dbBranch, nil
}

type BuildManager struct {
	collection *mgo.Collection
}

func NewBuildManager() *BuildManager {
	session := dbConf.Dbs.Copy()
	return &BuildManager{collection: session.DB(dbConf.Database).C("builds")}
}

func (manager *BuildManager) Insert(build *models.Build) error {

	err := manager.collection.Insert(build)
	if err != nil {
		return err
	}

	return nil
}

func (manager *BuildManager) FindByID(id string) ([]*models.Build, error) {

	var dbbuild []*models.Build
	err := manager.collection.Find(bson.M{"_id": id}).All(&dbbuild)
	if err != nil {
		return nil, err
	}

	return dbbuild, nil
}

func (manager *BuildManager) FindBuildByBranch(bid string) ([]*models.Build, error) {

	var dbbuild []*models.Build
	err := manager.collection.Find(bson.M{"branchid": bid}).All(&dbbuild)
	if err != nil {
		return nil, err
	}

	return dbbuild, nil
}
