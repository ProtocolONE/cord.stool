package controllers

import (
	"cord.stool/service/models"

	"encoding/json"
	"net/http"
)

func CreateCmd(w http.ResponseWriter, r *http.Request) {

	reqCmd := new(models.CreateCmd)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&reqCmd)

	w.Write([]byte("CreateCmd is running.\n"))
	w.Write([]byte("source: " + reqCmd.Source + "\n"))
	w.Write([]byte("output: " + reqCmd.Output + "\n"))
}

func PushCmd(w http.ResponseWriter, r *http.Request) {

	reqCmd := new(models.PushCmd)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&reqCmd)

	w.Write([]byte("PushCmd is running.\n"))
	w.Write([]byte("source: " + reqCmd.Source + "\n"))
	w.Write([]byte("output: " + reqCmd.Output + "\n"))
}

func DiffCmd(w http.ResponseWriter, r *http.Request) {

	reqCmd := new(models.DiffCmd)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&reqCmd)

	w.Write([]byte("DiffCmd is running.\n"))
	w.Write([]byte("old: " + reqCmd.SourceOld + "\n"))
	w.Write([]byte("new: " + reqCmd.SourceNew + "\n"))
	w.Write([]byte("patch: " + reqCmd.OutputDiff + "\n"))
}

func TorrentCmd(w http.ResponseWriter, r *http.Request) {

	reqCmd := new(models.TorrentCmd)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&reqCmd)

	w.Write([]byte("TorrentCmd API is running.\n"))
	w.Write([]byte("source: " + reqCmd.Source + "\n"))
	w.Write([]byte("target: " + reqCmd.Target + "\n"))
}
