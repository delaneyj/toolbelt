package zombiezen

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	f, err := os.OpenFile("sqlc-gen-zombiezen.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(f)
	log.Println("This is a test log entry")

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

	return res, nil
}
