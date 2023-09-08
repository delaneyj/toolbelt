package wisshes

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
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
	log.Print("Removing all files in artifacts directory")

	lastStatus := StepStatusUnchanged
	ctx = CtxWithInventory(ctx, inv)

	wg := &sync.WaitGroup{}
	wg.Add(len(inv.Hosts))
	errCh := make(chan error, len(inv.Hosts))

	for h, host := range inv.Hosts {
		hostName := inv.HostNames[h]
		go func(hostName string, host *goph.Client) {
			defer wg.Done()
			ctx = CtxWithSSHClient(ctx, host)
			ctx = CtxWithPreviousStep(ctx, StepStatusUnchanged)

			for i, step := range steps {
				log.Printf("[%s:%s] step %d started", hostName, host.Config.Addr, i+1)
				start := time.Now()
				var (
					name   string
					status StepStatus
					err    error
				)
				ctx, name, status, err = step(ctx)
				if err != nil {
					errCh <- fmt.Errorf("step: %w", err)
					return
				}

				if status == StepStatusFailed {
					errCh <- fmt.Errorf("step %d failed", i+1)
				}

				log.Printf("[%s:%s] step %d %s -> %s took %s", hostName, host.Config.Addr, i+1, name, status, time.Since(start))
				ctx = CtxWithPreviousStep(ctx, status)
				lastStatus = status
			}
		}(hostName, host)
	}
	wg.Wait()

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
