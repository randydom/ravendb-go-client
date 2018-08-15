package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTopologyTest_canGetTopology(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	command := NewGetDatabaseTopologyCommand()
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)

	assert.NotEqual(t, result.getEtag(), "")
	assert.Equal(t, len(result.getNodes()), 1)
	serverNode := result.getNodes()[0]
	assert.Equal(t, serverNode.GetUrl(), store.GetUrls()[0])
	assert.Equal(t, serverNode.GetDatabase(), store.GetDatabase())
	assert.Equal(t, serverNode.GetClusterTag(), "A")
	assert.Equal(t, serverNode.GetServerRole(), ServerNode_Role_MEMBER)
}

func TestGetTopology(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	getTopologyTest_canGetTopology(t)
}
