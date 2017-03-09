package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

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

func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)
	notFoundValue := reflect.Value{}
	if structFieldValue == notFoundValue {
		structFieldValue = structValue.FieldByNameFunc(
			func(fname string) bool {
				return strings.ToLower(name) == strings.ToLower(fname)
			})
	}

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		number, ok := value.(json.Number)
		if ok && structFieldType.Kind() == reflect.Uint32 {
			intValue, err := number.Int64()
			if err != nil {
				return err
			}
			structFieldValue.Set(reflect.ValueOf(uint32(intValue)))
			return nil
		}

		return errors.New("Provided value type didn't match obj field type")
	}

	structFieldValue.Set(val)
	return nil
}

func FillStruct(s interface{}, m map[string]interface{}) error {
	for k, v := range m {
		err := SetField(s, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
