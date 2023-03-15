package impl

import (
	"context"
	"encoding/json"
	"github.com/knqyf263/go-plugin/examples/host-function-library/library/json-parser/export"
)

var _ export.ParserLibrary = ParserLibraryImpl{}

// ParserLibraryImpl implements export.ParserLibrary functions
type ParserLibraryImpl struct{}

// ParseJson is embedded into the plugin and can be called by the plugin.
func (ParserLibraryImpl) ParseJson(_ context.Context, request *export.ParseJsonRequest) (*export.ParseJsonResponse, error) {
	var person export.Person
	if err := json.Unmarshal(request.GetContent(), &person); err != nil {
		return nil, err
	}

	return &export.ParseJsonResponse{Response: &person}, nil
}
