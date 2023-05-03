package internal

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2"
)

func (t *Tester) powerState(ctx context.Context, conn *bmclib.Client) (string, error) {
	return conn.GetPowerState(ctx)
}

func (t *Tester) powerSet(ctx context.Context, conn *bmclib.Client) (string, error) {
	return "", nil
}

func (t *Tester) userRead(ctx context.Context, conn *bmclib.Client) (string, error) {
	return "", nil
}

func (t *Tester) bmcReset(ctx context.Context, conn *bmclib.Client) (string, error) {
	return "", nil
}

func (t *Tester) bootDeviceSet(ctx context.Context, conn *bmclib.Client) (string, error) {
	return "", nil
}
