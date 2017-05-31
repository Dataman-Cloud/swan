package manager_test

import (
	//"flag"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	//"github.com/Dataman-Cloud/swan/src/config"
	//. "github.com/Dataman-Cloud/swan/src/manager"
	//"github.com/samuel/go-zookeeper/zk"
	//"github.com/urfave/cli"
)

func TestNew(t *testing.T) {
	// TODO(nmg): test will be finished later
	//set := flag.NewFlagSet("test", 0)
	//set.String("mesos", "zk://localhost:port1/mesos", "doc")
	//set.String("zk", "zk://localhost:port1/swan", "doc")
	//testCliCtx := cli.NewContext(nil, set, nil)
	//testManagerConf, _ := config.NewManagerConfig(testCliCtx)

	//testMgr, err := New(testManagerConf)
	//ok(t, err)

	//assert(t, testMgr != nil, "testMgr should be Manager instance")
}

// helper func
// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
