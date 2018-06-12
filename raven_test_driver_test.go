package ravendb

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

var (
	gRavenTestDriver *RavenTestDriver
)

type RavenTestDriver struct {
	globalServer               *DocumentStore
	globalServerProcess        *Process
	globalSecuredServer        *DocumentStore
	globalSecuredServerProcess *Process

	documentStores sync.Map // *DocumentStore => bool

	index    AtomicInteger
	disposed bool
}

func NewRavenTestDriver() *RavenTestDriver {
	return &RavenTestDriver{}
}

func (d *RavenTestDriver) getSecuredDocumentStore() (*DocumentStore, error) {
	return d.getDocumentStore2("test_db", true, 0)
}

// func (d *RavenTestDriver)
func (d *RavenTestDriver) getTestClientCertificate() *KeyStore {
	// TODO: implement me
	return nil
}

func (d *RavenTestDriver) getDocumentStore() (*DocumentStore, error) {
	return d.getDocumentStoreWithName("test_db")
}

func (d *RavenTestDriver) getSecuredDocumentStoreWithName(database string) (*DocumentStore, error) {
	return d.getDocumentStore2(database, true, 0)
}

func (d *RavenTestDriver) getDocumentStoreWithName(dbName string) (*DocumentStore, error) {
	return d.getDocumentStore2(dbName, false, 0)
}

func (d *RavenTestDriver) getDocumentStore2(dbName string, secured bool, waitForIndexingTimeout time.Duration) (*DocumentStore, error) {
	//fmt.Printf("getDocumentStore2\n")

	n := d.index.incrementAndGet()
	name := fmt.Sprintf("%s_%d", dbName, n)
	documentStore := d.getGlobalServer(secured)
	if documentStore == nil {
		err := d.runServer(secured)
		if err != nil {
			fmt.Printf("runServer failed with %s\n", err)
			return nil, err
		}
	}

	documentStore = d.getGlobalServer(secured)
	databaseRecord := NewDatabaseRecord()
	databaseRecord.DatabaseName = name

	createDatabaseOperation := NewCreateDatabaseOperation(databaseRecord)
	err := documentStore.maintenance().server().send(createDatabaseOperation)
	if err != nil {
		return nil, err
	}

	urls := documentStore.getUrls()
	store := NewDocumentStoreWithUrlsAndDatabase(urls, name)

	if secured {
		store.setCertificate(d.getTestClientCertificate())
	}

	// TODO: is over-written by CustomSerializationTest
	// customizeStore(store);
	d.hookLeakedConnectionCheck(store)

	d.setupDatabase(store)
	_, err = store.Initialize()
	if err != nil {
		return nil, err
	}

	fn := func(store *DocumentStore) {
		_, ok := d.documentStores.Load(store)
		if !ok {
			// TODO: shouldn't happen?
			return
		}

		operation := NewDeleteDatabasesOperation(store.getDatabase(), true)
		command := operation.getCommand(store.getConventions())
		store.maintenance().server().send(command)
	}

	store.addAfterCloseListener(fn)

	if waitForIndexingTimeout > 0 {
		d.waitForIndexing(store, name, waitForIndexingTimeout)
	}

	d.documentStores.Store(store, true)

	return store, nil
}

func (d *RavenTestDriver) hookLeakedConnectionCheck(store *DocumentStore) {
	// TODO: no-op for now. Not sure if I have enough info
	// to replicate this functionality in Go
}

// Note: it's virtual in Java but there's only one implementation
// that is a no-op
func (d *RavenTestDriver) setupDatabase(documentStore *DocumentStore) {
	// empty by design
}

func (d *RavenTestDriver) runServer(secured bool) error {
	var locator *RavenServerLocator
	var err error
	if secured {
		locator, err = NewSecuredServiceLocator()
	} else {
		locator, err = NewTestServiceLocator()
	}
	if err != nil {
		return err
	}
	fmt.Printf("runServer: starting server\n")
	proc, err := RavenServerRunner_run(locator)
	if err != nil {
		fmt.Printf("RavenServerRunner_run failed with %s\n", err)
		return err
	}
	d.setGlobalServerProcess(secured, proc)

	// parse stdout of the server to extract server listening port from line:
	// Server available on: http://127.0.0.1:50386
	wantedPrefix := "Server available on: "
	scanner := bufio.NewScanner(proc.stdoutReader)
	timeStart := time.Now()
	url := ""
	for scanner.Scan() {
		dur := time.Since(timeStart)
		if dur > time.Minute {
			break
		}
		s := scanner.Text()
		if RavenServerVerbose {
			fmt.Printf("%s\n", s)
		}
		if !strings.HasPrefix(s, wantedPrefix) {
			continue
		}
		s = strings.TrimPrefix(s, wantedPrefix)
		url = strings.TrimSpace(s)
		break
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	if url == "" {
		return fmt.Errorf("Unable to start server")
	}
	fmt.Printf("Server started on: '%s'\n", url)

	if RavenServerVerbose {
		go func() {
			_, err := io.Copy(os.Stdout, proc.stdoutReader)
			if !(err == nil || err == io.EOF) {
				fmt.Printf("io.Copy() failed with %s\n", err)
			}
		}()
	}

	time.Sleep(time.Second) // TODO: probably not necessary

	store := NewDocumentStore()
	store.setUrls([]string{url})
	store.setDatabase("test.manager")
	store.getConventions().setDisableTopologyUpdates(true)

	if secured {
		panicIf(true, "NYI")
		d.globalSecuredServer = store
		//TODO: KeyStore clientCert = getTestClientCertificate();
		//TODO: store.setCertificate(clientCert);
	} else {
		d.globalServer = store
	}
	_, err = store.Initialize()
	return err
}

func (d *RavenTestDriver) waitForIndexing(store *DocumentStore, database String, timeout time.Duration) {
	// TODO: implement me
	panicIf(true, "NYI")
}

func (d *RavenTestDriver) killGlobalServerProcess(secured bool) {
	if secured {
		if d.globalSecuredServerProcess != nil {
			d.globalSecuredServerProcess.cmd.Process.Kill()
			d.globalSecuredServerProcess = nil
		}
	} else {
		if d.globalServerProcess != nil {
			d.globalServerProcess.cmd.Process.Kill()
			d.globalServerProcess = nil
		}
	}
}

func (d *RavenTestDriver) getGlobalServer(secured bool) *DocumentStore {
	if secured {
		return d.globalSecuredServer
	}
	return d.globalServer
}

func (d *RavenTestDriver) setGlobalServerProcess(secured bool, p *Process) {
	if secured {
		d.globalSecuredServerProcess = p
	} else {
		d.globalServerProcess = p
	}
}

func (d *RavenTestDriver) close() {
	if d.disposed {
		return
	}

	fn := func(key, value interface{}) bool {
		documentStore := key.(*DocumentStore)
		documentStore.Close()
		return true
	}
	d.documentStores.Range(fn)
	d.disposed = true
}

var (
	useProxyCached *bool
)

func useProxy() bool {
	if useProxyCached != nil {
		return *useProxyCached
	}
	var b bool
	if os.Getenv("HTTP_PROXY") != "" {
		fmt.Printf("Using http proxy\n")
		b = true
	} else {
		fmt.Printf("Not using http proxy\n")
	}
	useProxyCached = &b
	return b
}

func shutdownTests() {
	gRavenTestDriver.killGlobalServerProcess(true)
	gRavenTestDriver.killGlobalServerProcess(false)
}

var dbTestsDisabledAlreadyPrinted = false

func dbTestsDisabled() bool {
	if os.Getenv("RAVEN_GO_NO_DB_TESTS") != "" {
		if !dbTestsDisabledAlreadyPrinted {
			dbTestsDisabledAlreadyPrinted = true
			fmt.Printf("DB tests are disabled\n")
		}
		return true
	}
	return false
}

func getDocumentStoreMust(t *testing.T) *DocumentStore {
	store, err := gRavenTestDriver.getDocumentStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)
	return store
}

func openSessionMust(t *testing.T, store *DocumentStore) *DocumentSession {
	session, err := store.OpenSession()
	assert.NoError(t, err)
	assert.NotNil(t, session)
	return session
}

func TestMain(m *testing.M) {
	noDb := os.Getenv("RAVEN_GO_NO_DB_TESTS")
	if noDb == "" {
		// this helps running tests from withing Visual Studio Code,
		// where env variables are not set
		serverPath := os.Getenv("RAVENDB_JAVA_TEST_SERVER_PATH")
		if serverPath == "" {
			home := os.Getenv("HOME")
			path := filepath.Join(home, "Documents", "RavenDB", "Server", "Raven.Server")
			_, err := os.Stat(path)
			if err != nil {
				cwd, err := os.Getwd()
				must(err)
				path = filepath.Join(cwd, "RavenDB", "Server", "Raven.Server")
				_, err = os.Stat(path)
				must(err)
			}
			os.Setenv("RAVENDB_JAVA_TEST_SERVER_PATH", path)
			fmt.Printf("Setting RAVENDB_JAVA_TEST_SERVER_PATH to '%s'\n", path)
		}
	}

	//RavenServerVerbose = true
	if useProxy() {
		go proxy.Run("")
	}
	gRavenTestDriver = NewRavenTestDriver()

	code := m.Run()

	// TODO: run this even when exception happens
	shutdownTests()

	if useProxy() {
		proxy.CloseLogFile()
		fmt.Printf("Closed proxy log file\n")
	}
	os.Exit(code)
}
