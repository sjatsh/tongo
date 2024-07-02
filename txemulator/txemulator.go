package txemulator

// #cgo darwin LDFLAGS: -L ../lib/darwin/ -Wl,-rpath,../lib/darwin/ -l emulator
// #cgo linux LDFLAGS: -L ../lib/linux/ -Wl,-rpath,../lib/linux/ -l emulator
// #cgo windows LDFLAGS: -L ../lib/windows/ -Wl,-rpath,../lib/windows/ -l emulator
// #include "../lib/emulator-extern.h"
// #include <stdlib.h>
// #include <stdbool.h>
import "C"
import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"unsafe"

	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
)

type VerbosityLevel int

const (
	LogTruncated VerbosityLevel = iota
	LogUnlimited
	CellHashAndOffsetForCommand
	PrintsAllStackValuesForCommand
)

type Emulator struct {
	emulator unsafe.Pointer
}

var DefaultConfig = `te6ccgECRwEABFcAAgmbAAAAEAECAgEgAwQCASAaGwIBIAUGAgEgDxACASAHCAEBYg4CASAJCgEBSA0BASALAQEgDABAVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVUAQDMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABA5WdU+DQm9psJJnvYdqyXxEghNFt+JmvZVqe/v7mN81wBAWIRAgEgEhMAGsQAAAACAAAAAAAAAC4BAUgUAgEgFhcBAcAVALfQUy7nTs8AAAJwACrYn7aHDoYaZOELB7fIx0lsFfzu58bxcmSlH++c6KojdwX2/yWZOw/Zr08OxAx1OQZWjQc9ppdrOeJEc5dIgaEAAAAAD/////gAAAAAAAAABAEBIBgBASAZABRrRlU/EAQ7msoAACAAAQAAAACAAAAAIAAAAIAAAgEgHB0CASAvMAIBIB4fAgEgJSYCASAgIQEBSCQBASAiAQEgIwAMAZAAZABLADdwEQ2TFuwAByOG8m/BAACAEKdBpGJ4AAAAMAAIAE3QZgAAAAAAAAAAAAAAAIAAAAAAAAD6AAAAAAAAAfQAAAAAAAPQkEACASAnKAIBICssAQEgKQEBICoAlNEAAAAAAAAAZAAAAAAAD0JA3gAAAAAnEAAAAAAAAAAPQkAAAAAAAhYOwAAAAAAAACcQAAAAAAI0k0AAAAAABfXhAAAAAAA7msoAAJTRAAAAAAAAAGQAAAAAAAGGoN4AAAAAA+gAAAAAAAAAD0JAAAAAAAAPQkAAAAAAAAAnEAAAAAAAmJaAAAAAAAX14QAAAAAAO5rKAAEBIC0BASAuAFBdwwACAAAACAAAABAAAMMAHoSAAB6EgAI0k0DDAAAD6AAAE4gAACcQAFBdwwACAAAACAAAABAAAMMAHoSAAJiWgAExLQDDAAAD6AAAE4gAACcQAgFIMTICASA1NgEBIDMBASA0AELqAAAAAACYloAAAAAAJxAAAAAAAA9CQAAAAAGAAFVVVVUAQuoAAAAAAA9CQAAAAAAD6AAAAAAAAYagAAAAAYAAVVVVVQIBIDc4AQFYOwEBIDkBASA6ACTCAQAAAPoAAAD6AAAD6AAAABcAStkBAwAAB9AAAD6AAAAAAwAAAAgAAAAEACAAAAAgAAAAAgAAJxABAcA8AgEgPT4CASA/QABDv+6SYlD5XEfFuCmona5jYtGN4iWVOW5abGAZxXh4ab9iwAIBIEFCAEK/jVwCELNdrdqiGfrEWdug/e+x+uTpeg0Hl3Of4FDWlMoCAUhDRAIBWEVGAAPfcABBvvXr/85ThwN08RVEkXrXOpCNTrUaVASnRwrD2wNe3bMUAEG+2ZmZmZmZmZmZmZmZmZmZmZmZmZmZmZmZmZmZmZmZmZwAQb7c3f6FapnFy4B4QZnAdwvqMfKODXM49zeESA3vRM2QFA==`

//	{
//	  "success": false,
//	  "error": "Error description"
//	  // and optional fields "vm_exit_code" and "vm_log" in case external message was not accepted.
//	}
//
// Or success:
//
//	{
//	  "success": true,
//	  "transaction": "Base64 encoded Transaction boc",
//	  "shard_account": "Base64 encoded new ShardAccount boc",
//	  "vm_log": "execute DUP...",
//	  "actions": "Base64 encoded compute phase actions boc (OutList n)"
//	}
type result struct {
	Success      bool   `json:"success"`
	Error        string `json:"error"`
	Transaction  string `json:"transaction"`
	ShardAccount string `json:"shard_account"`
	VmLog        string `json:"vm_log"`
	VmExitCode   int    `json:"vm_exit_code"`
	Actions      string `json:"actions"`
}

type EmulationResult struct {
	Success   bool
	Emulation *struct {
		ShardAccount   tlb.ShardAccount
		Transaction    tlb.Transaction
		RawTransaction string
	}
	Logs  string
	Error *struct {
		ExitCode int
		Text     string
	}
}

// NewEmulator
// Verbosity level of VM log. 0 - log truncated to last 256 characters. 1 - unlimited length log.
// 2 - for each command prints its cell hash and offset. 3 - for each command log prints all stack values.
func NewEmulator(config *boc.Cell, verbosityLevel VerbosityLevel) (*Emulator, error) {
	configBoc, err := config.ToBocBase64()
	if err != nil {
		return nil, err
	}
	return newEmulatorBase64(configBoc, verbosityLevel)
}

func newEmulatorBase64(configBoc string, verbosityLevel VerbosityLevel) (*Emulator, error) {

	cConfigStr := C.CString(configBoc)
	defer C.free(unsafe.Pointer(cConfigStr))
	level := C.int(verbosityLevel)
	e := Emulator{emulator: C.transaction_emulator_create(cConfigStr, level)}
	runtime.SetFinalizer(&e, destroy)
	return &e, nil
}

func (e *Emulator) SetUnixtime(utime uint32) error {
	ok := C.transaction_emulator_set_unixtime(e.emulator, C.uint32_t(utime))
	if !ok {
		return fmt.Errorf("set unixtime error")
	}
	return nil
}

func (e *Emulator) SetLT(lt uint64) error {
	ok := C.transaction_emulator_set_lt(e.emulator, C.uint64_t(lt))
	if !ok {
		return fmt.Errorf("set LT error")
	}
	return nil
}

func (e *Emulator) SetRandomSeed(seed [32]byte) error {
	cSeedStr := C.CString(hex.EncodeToString(seed[:]))
	defer C.free(unsafe.Pointer(cSeedStr))
	ok := C.transaction_emulator_set_rand_seed(e.emulator, cSeedStr)
	if !ok {
		return fmt.Errorf("set random seed error")
	}
	return nil
}

func (e *Emulator) SetIgnoreSignatureCheck(ignore bool) error {
	ok := C.transaction_emulator_set_ignore_chksig(e.emulator, C.bool(ignore))
	if !ok {
		return fmt.Errorf("set IgnoreSignatureCheck error")
	}
	return nil
}

func (e *Emulator) SetConfig(config *boc.Cell) error {
	configBoc, err := config.ToBocBase64()
	if err != nil {
		return err
	}
	cConfigStr := C.CString(configBoc)
	defer C.free(unsafe.Pointer(cConfigStr))
	ok := C.transaction_emulator_set_config(e.emulator, cConfigStr)
	if !ok {
		return fmt.Errorf("set config error")
	}
	return nil
}

func (e *Emulator) SetLibs(libs *boc.Cell) error {
	libsBoc, err := libs.ToBocBase64()
	if err != nil {
		return err
	}
	return e.setLibs(libsBoc)
}

func (e *Emulator) setLibs(libsBoc string) error {
	cLibsStr := C.CString(libsBoc)
	defer C.free(unsafe.Pointer(cLibsStr))
	ok := C.transaction_emulator_set_libs(e.emulator, cLibsStr)
	if !ok {
		return fmt.Errorf("set libs error")
	}
	return nil
}

func (e *Emulator) Emulate(shardAccount tlb.ShardAccount, message tlb.Message) (EmulationResult, error) {
	msg, err := tlbStructToBase64(message)
	if err != nil {
		return EmulationResult{}, err
	}

	acc, err := tlbStructToBase64(shardAccount)
	if err != nil {
		return EmulationResult{}, err
	}
	return e.emulateBase64(acc, msg)
}

func (e *Emulator) emulateBase64(acc string, msg string) (EmulationResult, error) {
	cAccStr := C.CString(acc)
	defer C.free(unsafe.Pointer(cAccStr))
	cMsgStr := C.CString(msg)
	defer C.free(unsafe.Pointer(cMsgStr))

	r := C.transaction_emulator_emulate_transaction(e.emulator, cAccStr, cMsgStr)
	rJSON := C.GoString(r)
	defer C.free(unsafe.Pointer(r))
	var (
		res     result
		account tlb.ShardAccount
		tx      tlb.Transaction
	)
	err := json.Unmarshal([]byte(rJSON), &res)
	if err != nil {
		return EmulationResult{}, err
	}

	if res.Success == false {
		err1 := struct {
			ExitCode int
			Text     string
		}{
			ExitCode: res.VmExitCode,
			Text:     res.Error,
		}
		return EmulationResult{Success: false, Logs: res.VmLog, Error: &err1}, nil
	}

	accountCell, err := boc.DeserializeBocBase64(res.ShardAccount)
	if err != nil {
		return EmulationResult{}, err
	}
	err = tlb.Unmarshal(accountCell[0], &account)
	if err != nil {
		return EmulationResult{}, err
	}

	txCell, err := boc.DeserializeBocBase64(res.Transaction)
	if err != nil {
		return EmulationResult{}, err
	}
	err = tlb.Unmarshal(txCell[0], &tx)
	if err != nil {
		return EmulationResult{}, err
	}
	em := struct {
		ShardAccount   tlb.ShardAccount
		Transaction    tlb.Transaction
		RawTransaction string
	}{
		ShardAccount:   account,
		Transaction:    tx,
		RawTransaction: res.Transaction,
	}
	return EmulationResult{Success: true, Logs: res.VmLog, Emulation: &em}, nil
}

func destroy(e *Emulator) {
	C.transaction_emulator_destroy(e.emulator)
}

// SetVerbosityLevel
// verbosity level (0 - never, 1 - error, 2 - warning, 3 - info, 4 - debug)
func (e *Emulator) SetVerbosityLevel(level int) error {
	ok := C.emulator_set_verbosity_level(C.int(level))
	if !ok {
		return fmt.Errorf("set VerbosityLevel error")
	}
	return nil
}

func tlbStructToBase64(s any) (string, error) {
	cell := boc.NewCell()
	err := tlb.Marshal(cell, s)
	if err != nil {
		return "", err
	}
	return cell.ToBocBase64()
}
