package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/twinj/uuid"
)

// LoadNodeID load or create id in local id file
func LoadNodeID(filePath string) (string, error) {
	filePath = filepath.Clean(filePath)

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(filePath), 0700)
		if err != nil {
			return "", err
		}
		nodeID := uuid.NewV4().String()
		return nodeID, ioutil.WriteFile(filePath, []byte(nodeID), os.FileMode(0400))
	}

	bs, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(bs)), nil
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
