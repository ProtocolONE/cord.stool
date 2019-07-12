package database

import (
	"cord.stool/service/config"
	"cord.stool/service/models"

	"fmt"
	"go.uber.org/zap"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
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

	timeout, _ := time.ParseDuration("30s")
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{cfg.Host},
		Database: cfg.Database,
		Username: cfg.User,
		Password: cfg.Password,
		Timeout:  timeout,
	})

	if err != nil {
		session, err := mgo.DialWithTimeout(cfg.Host, timeout)
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

func (manager *UserManager) Close() {

	manager.collection.Database.Session.Close()
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

func (manager *BranchManager) Close() {

	manager.collection.Database.Session.Close()
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

	err := manager.collection.RemoveId(id)
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

func (manager *BuildManager) Close() {

	manager.collection.Database.Session.Close()
}

func (manager *BuildManager) Insert(build *models.Build) error {

	err := manager.collection.Insert(build)
	if err != nil {
		return err
	}

	return nil
}

func (manager *BuildManager) RemoveByID(id string) error {

	err := manager.collection.RemoveId(id)
	if err != nil {
		return err
	}

	return nil
}

func (manager *BuildManager) FindByID(id string) (*models.Build, error) {

	var dbbuild []*models.Build
	err := manager.collection.Find(bson.M{"_id": id}).All(&dbbuild)
	if err != nil {
		return nil, err
	}

	if dbbuild == nil {
		return nil, nil
	}

	if len(dbbuild) > 1 {
		return nil, fmt.Errorf("Database integrity error")
	}

	return dbbuild[0], nil
}

func (manager *BuildManager) FindBuildByBranchID(bid string) ([]*models.Build, error) {

	var dbbuild []*models.Build
	err := manager.collection.Find(bson.M{"branchid": bid}).Sort("-created").All(&dbbuild)
	if err != nil {
		return nil, err
	}

	return dbbuild, nil
}

func (manager *BuildManager) FindLastBuildByBranchID(bid string) (*models.Build, error) {

	dbbuild, err := manager.FindBuildByBranchID(bid)
	if err != nil {
		return nil, err
	}

	if len(dbbuild) == 0 {
		return nil, nil
	}

	return dbbuild[0], nil
}

type DepotManager struct {
	collection *mgo.Collection
}

func NewDepotManager() *DepotManager {
	session := dbConf.Dbs.Copy()
	return &DepotManager{collection: session.DB(dbConf.Database).C("depots")}
}

func (manager *DepotManager) Close() {

	manager.collection.Database.Session.Close()
}

func (manager *DepotManager) Insert(depot *models.Depot) error {

	err := manager.collection.Insert(depot)
	if err != nil {
		return err
	}

	return nil
}

func (manager *DepotManager) RemoveByID(id string) error {

	err := manager.collection.RemoveId(id)
	if err != nil {
		return err
	}

	return nil
}

func (manager *DepotManager) FindByID(id string) (*models.Depot, error) {

	var dbdepot []*models.Depot
	err := manager.collection.Find(bson.M{"_id": id}).All(&dbdepot)
	if err != nil {
		return nil, err
	}

	if dbdepot == nil {
		return nil, nil
	}

	if len(dbdepot) > 1 {
		return nil, fmt.Errorf("Database integrity error")
	}

	return dbdepot[0], nil
}

type BuildDepotManager struct {
	collection *mgo.Collection
}

func NewBuildDepotManager() *BuildDepotManager {
	session := dbConf.Dbs.Copy()
	return &BuildDepotManager{collection: session.DB(dbConf.Database).C("build_depot")}
}

func (manager *BuildDepotManager) Close() {

	manager.collection.Database.Session.Close()
}

func (manager *BuildDepotManager) Insert(buildDepot *models.BuildDepot) error {

	err := manager.collection.Insert(buildDepot)
	if err != nil {
		return err
	}

	return nil
}

func (manager *BuildDepotManager) RemoveByID(id string) error {

	err := manager.collection.RemoveId(id)
	if err != nil {
		return err
	}

	return nil
}

func (manager *BuildDepotManager) Update(buildDepot *models.BuildDepot) error {

	err := manager.collection.UpdateId(buildDepot.ID, buildDepot)
	if err != nil {
		return err
	}

	return nil
}

func (manager *BuildDepotManager) FindByID(id string) (*models.BuildDepot, error) {

	var dbbuilddepot []*models.BuildDepot
	err := manager.collection.Find(bson.M{"_id": id}).All(&dbbuilddepot)
	if err != nil {
		return nil, err
	}

	if dbbuilddepot == nil {
		return nil, nil
	}

	if len(dbbuilddepot) > 1 {
		return nil, fmt.Errorf("Database integrity error")
	}

	return dbbuilddepot[0], nil
}

func (manager *BuildDepotManager) FindByBuildID(id string) ([]*models.BuildDepot, error) {

	var dbbuilddepot []*models.BuildDepot
	err := manager.collection.Find(bson.M{"buildid": id}).Sort("-created").All(&dbbuilddepot)
	if err != nil {
		return nil, err
	}

	return dbbuilddepot, nil
}

func (manager *BuildDepotManager) FindByBuildAndPlatformID(id string, platform string) (*models.BuildDepot, error) {

	var dbbuilddepot []*models.BuildDepot
	err := manager.collection.Find(bson.M{"buildid": id, "platform": platform}).Sort("-created").All(&dbbuilddepot)
	if err != nil {
		return nil, err
	}

	if dbbuilddepot == nil {
		return nil, nil
	}

	if len(dbbuilddepot) > 1 {
		return nil, fmt.Errorf("Database integrity error")
	}

	return dbbuilddepot[0], nil
}

type RedistrManager struct {
	collection *mgo.Collection
}

func NewRedistrManager() *RedistrManager {
	session := dbConf.Dbs.Copy()
	return &RedistrManager{collection: session.DB(dbConf.Database).C("redistr")}
}

func (manager *RedistrManager) Close() {

	manager.collection.Database.Session.Close()
}

func (manager *RedistrManager) FindByName(name string) (*models.Redistr, error) {

	var dbredistr []*models.Redistr
	err := manager.collection.Find(bson.M{"name": name}).All(&dbredistr)
	if err != nil {
		return nil, err
	}

	if dbredistr == nil {
		return nil, nil
	}

	if len(dbredistr) > 1 {
		return nil, fmt.Errorf("Database integrity error")
	}

	return dbredistr[0], nil
}
