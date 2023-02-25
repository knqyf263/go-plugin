package impl

import (
	"context"
	"encoding/json"

	"github.com/knqyf263/go-plugin/tests/import-host-functions/host/foo/export"
)

var _ export.ForeignHostFunctions = FooHostFunctions{}

// FooHostFunctions implements proto.ForeignHostFunctions
type FooHostFunctions struct{}

// ParseJson is embedded into the plugin and can be called by the plugin.
func (FooHostFunctions) ParseJson(ctx context.Context, request export.ParseJsonRequest) (export.ParseJsonResponse, error) {
	var person export.Person
	if err := json.Unmarshal(request.GetContent(), &person); err != nil {
		return export.ParseJsonResponse{}, err
	}

	return export.ParseJsonResponse{Response: &person}, nil
}
