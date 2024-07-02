package tlb

import (
	"github.com/sjatsh/tongo/boc"
)

// ShardAccount
// account_descr$_ account:^Account last_trans_hash:bits256
// last_trans_lt:uint64 = ShardAccount;
type ShardAccount struct {
	Account       Account `tlb:"^"`
	LastTransHash Bits256
	LastTransLt   uint64
}

// Account
// account_none$0 = Account;
type Account struct {
	SumType
	AccountNone struct {
	} `tlbSumType:"account_none$0"`
	Account ExistedAccount `tlbSumType:"account$1"`
}

// account$1 addr:MsgAddressInt storage_stat:StorageInfo
// storage:AccountStorage = Account;
type ExistedAccount struct {
	Addr        MsgAddress
	StorageStat StorageInfo
	Storage     AccountStorage
}

func (a Account) CurrencyCollection() (CurrencyCollection, bool) {
	switch a.SumType {
	case "AccountNone":
		return CurrencyCollection{}, true
	case "Account":
		return a.Account.Storage.Balance, true
	default:
		return CurrencyCollection{}, false
	}
}

func (a Account) Status() AccountStatus {
	if a.SumType == "AccountNone" {
		return AccountNone
	}
	switch a.Account.Storage.State.SumType {
	case "AccountUninit":
		return AccountUninit
	case "AccountActive":
		return AccountActive
	case "AccountFrozen":
		return AccountFrozen
	}
	panic("invalid sum types for account status")
}

// AccountStorage
// account_storage$_ last_trans_lt:uint64
// balance:CurrencyCollection state:AccountState
// = AccountStorage;
type AccountStorage struct {
	LastTransLt uint64
	Balance     CurrencyCollection
	State       AccountState
}

// AccountState
// account_uninit$00 = AccountState;
// account_active$1 _:StateInit = AccountState;
// account_frozen$01 state_hash:bits256 = AccountState;
type AccountState struct {
	SumType
	AccountUninit struct {
	} `tlbSumType:"account_uninit$00"`
	AccountActive struct {
		StateInit StateInit
	} `tlbSumType:"account_active$1"`
	AccountFrozen struct {
		StateHash Bits256
	} `tlbSumType:"account_frozen$01"`
}

// StorageInfo
// storage_info$_ used:StorageUsed last_paid:uint32
// due_payment:(Maybe Grams) = StorageInfo;
type StorageInfo struct {
	Used       StorageUsed
	LastPaid   uint32
	DuePayment Maybe[Grams]
}

// StorageUsed
// storage_used$_ cells:(VarUInteger 7) bits:(VarUInteger 7)
// public_cells:(VarUInteger 7) = StorageUsed;
type StorageUsed struct {
	Cells       VarUInteger7
	Bits        VarUInteger7
	PublicCells VarUInteger7
}

// AccountStatus
// acc_state_uninit$00 = AccountStatus;
// acc_state_frozen$01 = AccountStatus;
// acc_state_active$10 = AccountStatus;
// acc_state_nonexist$11 = AccountStatus;
type AccountStatus string

const (
	//AccountEmpty  AccountStatus = "empty" // empty state from node
	AccountNone   AccountStatus = "nonexist"
	AccountUninit AccountStatus = "uninit"
	AccountActive AccountStatus = "active"
	AccountFrozen AccountStatus = "frozen"
)

func (a AccountStatus) MarshalTLB(c *boc.Cell, encoder *Encoder) error {
	switch a {
	case AccountUninit:
		return c.WriteUint(0, 2)
	case AccountFrozen:
		return c.WriteUint(1, 2)
	case AccountActive:
		return c.WriteUint(2, 2)
	case AccountNone:
		return c.WriteUint(3, 2)
	}
	return nil
}

func (a *AccountStatus) UnmarshalTLB(c *boc.Cell, decoder *Decoder) error {
	t, err := c.ReadUint(2)
	if err != nil {
		return err
	}
	switch t {
	case 0:
		*a = AccountUninit
	case 1:
		*a = AccountFrozen
	case 2:
		*a = AccountActive
	case 3:
		*a = AccountNone
	}
	return nil
}
