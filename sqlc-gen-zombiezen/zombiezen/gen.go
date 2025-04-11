package zombiezen

import (
	"context"
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	res := &plugin.GenerateResponse{}

	querieFiles, err := generateQueries(req)
	if err != nil {
		return nil, fmt.Errorf("generating queries: %w", err)
	}
	res.Files = append(res.Files, querieFiles...)

	crudFiles, err := generateCRUD(req)
	if err != nil {
		return nil, fmt.Errorf("generating crud: %w", err)
	}
	res.Files = append(res.Files, crudFiles...)

	utilFiles, err := generateUtil(req)
	if err != nil {
		return nil, fmt.Errorf("generating crud: %w", err)
	}
	res.Files = append(res.Files, utilFiles...)

	return res, nil
}
