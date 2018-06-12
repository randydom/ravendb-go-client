package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type GetDatabaseNamesOperation struct {
	_start    int
	_pageSize int
}

func NewGetDatabaseNamesOperation(_start int, _pageSize int) *GetDatabaseNamesOperation {
	return &GetDatabaseNamesOperation{
		_start:    _start,
		_pageSize: _pageSize,
	}
}

func (o *GetDatabaseNamesOperation) getCommand(conventions *DocumentConventions) *GetDatabaseNamesCommand {
	return NewGetDatabaseNamesCommand(o._start, o._pageSize)
}

var (
	_ RavenCommand = &GetDatabaseNamesCommand{}
)

type GetDatabaseNamesCommand struct {
	*RavenCommandBase

	_start    int
	_pageSize int

	Result []string
}

func NewGetDatabaseNamesCommand(_start int, _pageSize int) *GetDatabaseNamesCommand {
	cmd := &GetDatabaseNamesCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_start:    _start,
		_pageSize: _pageSize,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetDatabaseNamesCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases?start=" + strconv.Itoa(c._start) + "&pageSize=" + strconv.Itoa(c._pageSize) + "&namesOnly=true"

	return NewHttpGet(url)
}

// GetDatabaseNamesResult describes response of GetDatabaseNames command
type GetDatabaseNamesResult struct {
	Databases []string `json:"Databases"`
}

func (c *GetDatabaseNamesCommand) setResponse(response String, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}

	var res GetDatabaseNamesResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}

	/*
		JsonNode names = mapper.readTree(response);
		if (!names.has("Databases")) {
			throwInvalidResponse();
		}

		JsonNode databases = names.get("Databases");
		if (!databases.isArray()) {
			return throwInvalidResponse();
		}
		ArrayNode dbNames = (ArrayNode) databases;
		String[] databaseNames = new String[dbNames.size()];
		for (int i = 0; i < dbNames.size(); i++) {
			databaseNames[i] = dbNames.get(i).asText();
		}
	*/

	c.Result = res.Databases
	return nil
}
