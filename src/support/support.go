package support

import (
	"context"

	"github.com/zeropsio/zcli/src/uuid"
)

type supportID struct{}

const Key = "VshApp-Chain-Id"

func Context(ctx context.Context) context.Context {
	id := uuid.GetShort()
	return context.WithValue(ctx, supportID{}, id)
}

func GetID(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(supportID{}).(string)
	return key, ok
}
