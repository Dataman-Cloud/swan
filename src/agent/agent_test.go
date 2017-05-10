package agent_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	. "github.com/Dataman-Cloud/swan/src/agent"
	"github.com/Dataman-Cloud/swan/src/config"
	// "golang.org/x/net/context"
)

func TestStartAndJoin(t *testing.T) {
	testConf := config.AgentConfig{ListenAddr: "localhost:8881", GossipListenAddr: "localhost:8881", GossipJoinAddr: "localhost:8881"}
	testAgent, err := New(testConf)

	ok(t, err)
	assert(t, testAgent != nil, "agent should be Agent instance")

	resolvCfg := &nameserver.Config{
		ListenAddr: "localhost:8882",
	}
	testAgent.Resolver = nameserver.NewResolver(resolvCfg)

	// TODO need mock agent.start
	// err1 := testAgent.StartAndJoin(context.TODO())
	// ok(t, err1)
}

// harness testing func
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
