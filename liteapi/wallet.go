package liteapi

import (
	"context"
	"fmt"

	"github.com/sjatsh/tongo/tlb"
	"github.com/sjatsh/tongo/ton"
)

func (c *Client) GetSeqno(ctx context.Context, account ton.AccountID) (uint32, error) {
	errCode, stack, err := c.RunSmcMethod(ctx, account, "seqno", tlb.VmStack{})
	if err != nil {
		return 0, err
	}
	if errCode == 0xFFFFFF00 {
		return 0, nil
	} else if errCode != 0 && errCode != 1 {
		return 0, fmt.Errorf("method execution failed with code: %v", errCode)
	}
	if len(stack) != 1 || stack[0].SumType != "VmStkTinyInt" {
		return 0, fmt.Errorf("invalid stack")
	}
	return uint32(stack[0].VmStkTinyInt), nil
}
