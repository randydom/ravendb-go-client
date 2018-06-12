package ravendb

import (
	"fmt"
	"net/http"
)

var (
	_ RavenCommand = &HeadDocumentCommand{}
)

type HeadDocumentCommand struct {
	*RavenCommandBase
	_id           string
	_changeVector *string

	Result *string // change vector
}

func NewHeadDocumentCommand(id string, changeVector *string) *HeadDocumentCommand {
	panicIf(id == "", "id cannot be empty")
	cmd := &HeadDocumentCommand{
		RavenCommandBase: NewRavenCommandBase(),
		_id:              id,
		_changeVector:    changeVector,
	}

	return cmd
}

func (c *HeadDocumentCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?id=" + UrlUtils_escapeDataString(c._id)

	request, err := NewHttpHead(url)
	if err != nil {
		return nil, err
	}

	if c._changeVector != nil {
		request.Header.Set("If-None-Match", *c._changeVector)
	}

	return request, nil
}

func (c *HeadDocumentCommand) processResponse(cache *HttpCache, response *http.Response, url String) (ResponseDisposeHandling, error) {
	statusCode := response.StatusCode
	if statusCode == http.StatusNotModified {
		c.Result = c._changeVector
		return ResponseDisposeHandling_AUTOMATIC, nil
	}

	if statusCode == http.StatusNotFound {
		c.Result = nil
		return ResponseDisposeHandling_AUTOMATIC, nil
	}

	var err error
	c.Result, err = HttpExtensions_getRequiredEtagHeader(response)
	return ResponseDisposeHandling_AUTOMATIC, err
}

func (c *HeadDocumentCommand) setResponse(response String, fromCache bool) error {
	if response != "" {
		return throwInvalidResponse()
	}
	// TODO: is this really ever reached?
	fmt.Printf("HeadDocumentCommand.setResponse: clearing Result\n")
	c.Result = nil
	return nil
}
