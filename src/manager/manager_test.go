package manager

import (
	"os"
	"testing"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	tempFileDir := os.TempDir()
	defer func() {
		os.RemoveAll(tempFileDir)
	}()

	managerConfig := config.ManagerConfig{
		DataDir: tempFileDir,
	}

	m, err := New("foobar", managerConfig)
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

func TestLoadNode(t *testing.T) {
}
