package wisshes

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/zeebo/xxh3"
)

func FileRawToRemote(remotePath string, contents []byte) Step {
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
		log.Printf("Downloading %s to %s", remotePath, localPath)
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
			log.Printf("File %s unchanged", remotePath)
			return ctx, name, StepStatusUnchanged, err
		}
		log.Printf("File %s changed", remotePath)

		if err := os.WriteFile(localPath, contents, 0644); err != nil {
			return ctx, name, StepStatusFailed, fmt.Errorf("write file: %w", err)
		}

		remoteDir := filepath.Dir(remotePath)
		if _, err := RunF(client, "mkdir -p %s", remoteDir); err != nil {
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

		log.Printf("File %s updated", remotePath)

		return ctx, name, StepStatusChanged, nil
	}
}

func FilepathToRemote(remotePath string, localPath string) Step {
	name := fmt.Sprintf("file-remote-%s", strcase.ToKebab(remotePath))
	return func(ctx context.Context) (context.Context, string, StepStatus, error) {
		client := CtxSSHClient(ctx)

		sftp, err := client.NewSftp()
		if err != nil {
			return ctx, name, StepStatusFailed, err
		}
		defer sftp.Close()

		remoteDir := filepath.Dir(remotePath)
		if _, err := RunF(client, "mkdir -p %s", remoteDir); err != nil {
			return ctx, name, StepStatusFailed, fmt.Errorf("mkdir: %w", err)
		}

		remoteFile, err := sftp.Create(remotePath)
		if err != nil {
			return ctx, name, StepStatusFailed, fmt.Errorf("create: %w", err)
		}
		defer remoteFile.Close()

		if err := client.Upload(localPath, remotePath); err != nil {
			return ctx, name, StepStatusFailed, fmt.Errorf("upload: %w", err)
		}

		log.Printf("File %s updated", remotePath)

		return ctx, name, StepStatusChanged, nil
	}
}
