package linode

import (
	"context"
	"fmt"
	"net/http"

	"github.com/delaneyj/toolbelt/wisshes"
	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

const (
	ctxLinodeKeyClient  = ctxLinodeKeyPrefix + "client"
	ctxLinodeKeyAccount = ctxLinodeKeyPrefix + "account"
)

func CtxLinodeClient(ctx context.Context) *linodego.Client {
	return ctx.Value(ctxLinodeKeyClient).(*linodego.Client)
}

func CtxWithLinodeClient(ctx context.Context, client *linodego.Client) context.Context {
	return context.WithValue(ctx, ctxLinodeKeyClient, client)
}

func CtxLinodeAccount(ctx context.Context) *linodego.Account {
	return ctx.Value(ctxLinodeKeyAccount).(*linodego.Account)
}

func CtxWithLinodeAccount(ctx context.Context, account *linodego.Account) context.Context {
	return context.WithValue(ctx, ctxLinodeKeyAccount, account)
}

func ClientAndAccount(token string) wisshes.Step {

	return func(ctx context.Context) (context.Context, string, wisshes.StepStatus, error) {
		name := "linode client and account"

		linodeClient, acc, err := ClientFromToken(ctx, token)
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("failed to get linode client and account: %w", err)
		}

		ctx = CtxWithLinodeClient(ctx, linodeClient)
		ctx = CtxWithLinodeAccount(ctx, acc)
		return ctx, name, wisshes.StepStatusUnchanged, nil
	}
}

func ClientFromToken(ctx context.Context, token string) (*linodego.Client, *linodego.Account, error) {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	linodeClient := linodego.NewClient(oauth2Client)
	//linodeClient.SetDebug(true)

	acc, err := linodeClient.GetAccount(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &linodeClient, acc, nil

}
