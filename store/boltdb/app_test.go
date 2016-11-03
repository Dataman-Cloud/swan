package boltdb

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestAppCURD(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "swan")
	assert.Nil(t, err)

	defer os.Remove(tmpFile.Name())

	db, err := bolt.Open(tmpFile.Name(), 0644, nil)
	assert.Nil(t, err)

	boltdb := NewBoltdbStore(db)
	assert.Nil(t, err)

	app1 := &types.Application{
		ID:   "111111",
		Name: "test11111",
	}

	app2 := &types.Application{
		ID:   "222222",
		Name: "test222222",
	}

	err = boltdb.PutApps([]*types.Application{app1, app2}...)
	assert.Nil(t, err)

	tx, err := boltdb.Begin(false)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	bkt := getAppsBucket(tx)

	err = bkt.ForEach(func(k, v []byte) error {
		fmt.Println(string(k))
		var app types.Application
		if err := proto.Unmarshal(v, &app); err != nil {
			return err
		}

		fmt.Println(app.ID)
		fmt.Printf("app %+v \n", app)

		return nil
	})

	assert.Nil(t, err)

	boltdb.View(func(tx *bolt.Tx) error {
		b := getAppBucket(tx, app1.ID)
		b.ForEach(func(k, v []byte) error {
			fmt.Println(string(k))
			var app types.Application
			if err := proto.Unmarshal(v, &app); err != nil {
				return err
			}

			fmt.Println(v)
			fmt.Printf("app %+v \n", app)

			return nil
		})
		return nil

	})

	gotApp1, err := boltdb.GetApps(app1.ID)
	assert.Nil(t, err)
	fmt.Println(gotApp1[0].ID, gotApp1[0].Name)

	allApps, err := boltdb.GetApps()
	assert.Nil(t, err)
	fmt.Printf("apps:  %#v", allApps)
}
