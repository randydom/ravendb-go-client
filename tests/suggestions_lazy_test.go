package tests

import (
	"reflect"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func suggestionsLazy_usingLinq(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	indexDefinition := ravendb.NewIndexDefinition()
	indexDefinition.Name = "test"

	indexDefinition.Maps = []string{"from doc in docs.Users select new { doc.name }"}
	indexFieldOptions := ravendb.NewIndexFieldOptions()
	indexFieldOptions.Suggestions = true
	indexDefinition.Fields["name"] = indexFieldOptions

	err = store.Maintenance().Send(ravendb.NewPutIndexesOperation(indexDefinition))
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		user1 := User4{
			Name: "Ayende",
		}
		err = session.Store(user1)
		assert.NoError(t, err)

		user2 := User4{
			Name: "Oren",
		}
		err = session.Store(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	gRavenTestDriver.waitForIndexing(store, "", 0)

	{
		s := openSessionMust(t, store)
		oldRequests := s.Advanced().GetNumberOfRequests()

		q := s.QueryWithQueryOld(reflect.TypeOf(&User4{}), ravendb.Query_index("test"))
		fn := func(x ravendb.ISuggestionBuilder) {
			x.ByField("name", "Oren")
		}
		q2 := q.SuggestUsingBuilder(fn)
		suggestionQueryResult := q2.ExecuteLazy(nil)
		assert.Equal(t, oldRequests, s.Advanced().GetNumberOfRequests())

		resultI, err := suggestionQueryResult.GetValue()
		assert.NoError(t, err)
		result := resultI.(map[string]*ravendb.SuggestionResult)
		suggestions := result["name"].Suggestions
		assert.Equal(t, len(suggestions), 1)
		assert.Equal(t, suggestions[0], "oren")

		assert.Equal(t, oldRequests+1, s.Advanced().GetNumberOfRequests())

	}
}

func TestSuggestionsLazy(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	//suggestionsLazy_usingLinq(t)
}