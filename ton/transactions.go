package ton

import "github.com/sjatsh/tongo/tlb"

type Transaction struct {
	tlb.Transaction
	BlockID BlockIDExt
}
