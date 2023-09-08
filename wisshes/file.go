package wisshes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/zeebo/xxh3"
)

func FileRemote(remotePath string, contents []byte) Step {
	name := fmt.Sprintf("file-remote-%s", strcase.ToKebab(remotePath))
	return func(ctx context.Context) (context.Context, string, StepStatus, error) {
		client := CtxSSHClient(ctx)
		inv := CtxInventory(ctx)

		sftp, err := client.NewSftp()
		if err != nil {
			return ctx, name, StepStatusFailed, err
		}
		defer sftp.Close()

		localPath := inv.createTmpFilepath()
		if err := client.Download(remotePath, localPath); err != nil {
			if !os.IsNotExist(err) {
				return ctx, name, StepStatusFailed, fmt.Errorf("download: %w", err)
			}
		}
		b, err := os.ReadFile(localPath)
		if err != nil {
			return ctx, name, StepStatusFailed, fmt.Errorf("read file: %w", err)
		}
		remoteHash := xxh3.Hash(b)
		localHash := xxh3.Hash(contents)

		if remoteHash == localHash {
			return ctx, name, StepStatusUnchanged, err
		}

		if err := os.WriteFile(localPath, contents, 0644); err != nil {
			return ctx, name, StepStatusFailed, fmt.Errorf("write file: %w", err)
		}

		remoteDir := filepath.Dir(remotePath)
		if _, err := RunFn(client, "mkdir -p %s", remoteDir); err != nil {
			return ctx, name, StepStatusFailed, fmt.Errorf("mkdir: %w", err)
		}

		remoteFile, err := sftp.Create(remotePath)
		if err != nil {
			return ctx, name, StepStatusFailed, fmt.Errorf("create: %w", err)
		}
		defer remoteFile.Close()

		if _, err := remoteFile.Write(contents); err != nil {
			return ctx, name, StepStatusFailed, fmt.Errorf("copy: %w", err)
		}

		return ctx, name, StepStatusChanged, nil
	}
}
