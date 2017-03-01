package utils

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/twinj/uuid"
)

func LoadNodeID(filePath string) (string, error) {
	if !fileutil.Exist(filePath) {
		return "", errors.New("ID file was not found")
	}

	idFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	nodeID, err := ioutil.ReadAll(idFile)
	if err != nil {
		return "", err
	}

	logrus.Infof("starting swan node, ID file was found started with ID: %s", string(nodeID))

	return string(nodeID), nil
}

func CreateNodeID(filePath string) (string, error) {
	nodeID := uuid.NewV4().String()
	idFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}

	if _, err = idFile.WriteString(nodeID); err != nil {
		return "", err
	}

	logrus.Infof("starting swan node, ID file was not found started with  new ID: %s", nodeID)

	return nodeID, nil
}
