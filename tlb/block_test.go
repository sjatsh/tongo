package tlb

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"testing"

	"github.com/sjatsh/tongo/boc"
)

func Test_tlb_Unmarshal(t *testing.T) {
	type Transaction struct {
		AccountAddr   string
		Lt            uint64
		PrevTransHash string
		PrevTransLt   uint64
		Now           uint32
		OutMsgCnt     Uint15
		OrigStatus    AccountStatus
		EndStatus     AccountStatus
	}
	type AccountBlock struct {
		Transactions map[uint64]Transaction
	}
	type BlockContent struct {
		Accounts          map[string]*AccountBlock
		TxHashes          []string
		ValueFlow         ValueFlow
		InMsgDescrLength  int
		OutMsgDescrLength int
		Libraries         map[string]map[string]struct{}
	}
	testCases := []struct {
		name   string
		folder string
	}{
		{
			name:   "block (0,8000000000000000,30816553)",
			folder: "testdata/block-1",
		},
		{
			name:   "block (0,8000000000000000,40484416)",
			folder: "testdata/block-2",
		},
		{
			name:   "block (0,8000000000000000,40484438)",
			folder: "testdata/block-3",
		},
		{
			name:   "block (0,D83800000000000,4168601)",
			folder: "testdata/block-4",
		},
		{
			name:   "block (0,D83800000000000,(-1,8000000000000000,17734191)",
			folder: "testdata/block-5",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			inputFilename := path.Join(tc.folder, "block.bin")
			data, err := os.ReadFile(inputFilename)
			if err != nil {
				t.Fatalf("ReadFile() failed: %v", err)
			}
			cell, err := boc.DeserializeBoc(data)
			if err != nil {
				t.Fatalf("boc.DeserializeBoc() failed: %v", err)
			}
			var block Block
			err = Unmarshal(cell[0], &block)
			if err != nil {
				t.Fatalf("Unmarshal() failed: %v", err)
			}
			accounts := map[string]*AccountBlock{}
			var txHashes []string
			for _, account := range block.Extra.AccountBlocks.Values() {
				accBlock, ok := accounts[hex.EncodeToString(account.AccountAddr[:])]
				if !ok {
					accBlock = &AccountBlock{
						Transactions: map[uint64]Transaction{},
					}
					accounts[hex.EncodeToString(account.AccountAddr[:])] = accBlock
				}
				for _, txRef := range account.Transactions.Values() {
					tx := txRef.Value
					accBlock.Transactions[txRef.Value.Lt] = Transaction{
						AccountAddr:   hex.EncodeToString(tx.AccountAddr[:]),
						Lt:            tx.Lt,
						PrevTransHash: hex.EncodeToString(tx.PrevTransHash[:]),
						PrevTransLt:   tx.PrevTransLt,
						Now:           tx.Now,
						OutMsgCnt:     tx.OutMsgCnt,
						OrigStatus:    tx.OrigStatus,
						EndStatus:     tx.EndStatus,
					}
					txHashes = append(txHashes, tx.Hash().Hex())
				}
			}
			sort.Slice(txHashes, func(i, j int) bool {
				return txHashes[i] < txHashes[j]
			})
			inMsgLength, err := block.Extra.InMsgDescrLength()
			if err != nil {
				t.Errorf("InMsgDescrLength() failed: %v", err)
			}
			outMsgLength, err := block.Extra.OutMsgDescrLength()
			if err != nil {
				t.Errorf("InMsgDescrLength() failed: %v", err)
			}
			blk := BlockContent{
				Accounts:          accounts,
				TxHashes:          txHashes,
				ValueFlow:         block.ValueFlow,
				InMsgDescrLength:  inMsgLength,
				OutMsgDescrLength: outMsgLength,
				Libraries:         libraries(&block.StateUpdate.ToRoot),
			}
			bs, err := json.MarshalIndent(blk, " ", "  ")
			if err != nil {
				t.Errorf("json.MarshalIndent() failed: %v", err)
			}
			outputFilename := path.Join(tc.folder, "block.output.json")
			if err := os.WriteFile(outputFilename, bs, 0644); err != nil {
				t.Errorf("WriteFile() failed: %v", err)
			}
			expectedFilename := path.Join(tc.folder, "block.expected.json")
			content, err := os.ReadFile(expectedFilename)
			if err != nil {
				t.Errorf("ReadFile() failed: %v", err)
			}
			if bytes.Compare(bytes.Trim(content, " \n"), bytes.Trim(bs, " \n")) != 0 {
				t.Errorf("block content mismatch")
			}
		})
	}
}

func libraries(s *ShardState) map[string]map[string]struct{} {
	libs := map[string]map[string]struct{}{}
	for _, item := range s.UnsplitState.Value.ShardStateUnsplit.Other.Libraries.Items() {
		lib := fmt.Sprintf("%x", item.Key)
		if _, ok := libs[lib]; !ok {
			libs[lib] = map[string]struct{}{}
		}
		for _, pub := range item.Value.Publishers.Keys() {
			libs[lib][fmt.Sprintf("%x", pub)] = struct{}{}
		}
	}
	return libs
}
