package wisshes

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autosegment/ksuid"
	"github.com/melbahja/goph"
)

var WishDir = ".wisshes"

type Inventory struct {
	Hosts     []*goph.Client
	HostNames []string
}

func NewInventory(rootPassword string, namesAndIPs ...string) (inv *Inventory, err error) {

	if len(namesAndIPs) == 0 {
		return nil, fmt.Errorf("namesAndIPs is empty")
	}
	if len(namesAndIPs)%2 != 0 {
		return nil, fmt.Errorf("namesAndIPs must be even")
	}

	inv = &Inventory{}

	for i := 0; i < len(namesAndIPs); i += 2 {
		name := namesAndIPs[i]
		ip := namesAndIPs[i+1]
		inv.HostNames = append(inv.HostNames, name)

		host, err := goph.NewUnknown("root", ip, goph.Password(rootPassword))
		if err != nil {
			return nil, fmt.Errorf("new unknown: %w", err)
		}
		inv.Hosts = append(inv.Hosts, host)
	}

	if err := upsertWishDir(); err != nil {
		return nil, fmt.Errorf("upsert wish dir: %w", err)
	}

	return inv, nil
}

func (inv *Inventory) createTmpFilepath() string {
	return filepath.Join(TempDir(), ksuid.New().String())
}

func TempDir() string {
	return filepath.Join(WishDir, "tmp")
}

func ArtifactsDir() string {
	return filepath.Join(WishDir, "artifacts")
}

func upsertWishDir() error {
	if err := os.MkdirAll(ArtifactsDir(), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := os.RemoveAll(TempDir()); err != nil {
		return fmt.Errorf("remove all: %w", err)
	}
	if err := os.MkdirAll(TempDir(), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return nil
}

func (inv *Inventory) Close() {
	for _, host := range inv.Hosts {
		host.Close()
	}
}

func (inv *Inventory) Run(ctx context.Context, steps ...Step) (StepStatus, error) {
	lastStatus := StepStatusUnchanged
	ctx = CtxWithInventory(ctx, inv)

	if len(steps) == 0 {
		return lastStatus, nil
	}

	var (
		name   string
		status StepStatus
		err    error
	)

	for h, host := range inv.Hosts {
		hostName := inv.HostNames[h]
		if strings.Contains(hostName, "us-east") {
			log.Print("us-east")
		}
		ctx = CtxWithSSHClient(ctx, host)
		ctx = CtxWithPreviousStep(ctx, StepStatusUnchanged)

		for i, step := range steps {
			log.Printf("[%s:%s] step %d started", hostName, host.Config.Addr, i+1)
			start := time.Now()

			ctx, name, status, err = step(ctx)
			if err != nil {
				return status, fmt.Errorf("step %d: %w", i+1, err)
			}

			if status == StepStatusFailed {
				return status, fmt.Errorf("step %d: %w", i+1, err)
			}

			log.Printf("[%s:%s] step %d %s -> %s took %s", hostName, host.Config.Addr, i+1, name, status, time.Since(start))
			ctx = CtxWithPreviousStep(ctx, status)
			lastStatus = status
		}
	}

	return lastStatus, nil
}

func Run(ctx context.Context, steps ...Step) error {
	if err := upsertWishDir(); err != nil {
		return fmt.Errorf("upsert wish dir: %w", err)
	}

	if _, _, _, err := RunAll(steps...)(ctx); err != nil {
		return fmt.Errorf("run all: %w", err)
	}
	return nil
}
