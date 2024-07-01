package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/delaneyj/toolbelt/sqlc-gen-zombiezen/pb/plugin"
	"github.com/delaneyj/toolbelt/sqlc-gen-zombiezen/zombiezen"
	"google.golang.org/protobuf/proto"
)

func main() {

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error generating JSON: %s", err)
		os.Exit(2)
	}
}

func run() error {
	reqBlob, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read request: %w", err)
	}
	req := &plugin.GenerateRequest{}
	if err := proto.Unmarshal(reqBlob, req); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	resp, err := zombiezen.Generate(context.Background(), req)
	if err != nil {
		return err
	}

	// if usesStdin {
	respBlob, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(os.Stdout)
	if _, err := w.Write(respBlob); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}
