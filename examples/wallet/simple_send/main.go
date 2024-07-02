package main

import (
	"context"
	"github.com/sjatsh/tongo"
	"github.com/sjatsh/tongo/liteapi"
	"github.com/sjatsh/tongo/ton"
	"github.com/sjatsh/tongo/wallet"
	"log"
)

const SEED = "best journey rifle scheme bamboo daring finish life have puzzle verb wagon double pencil plate parent canoe soup stable salon drift elephant border hero"

func main() {
	client, err := liteapi.NewClientWithDefaultMainnet()
	if err != nil {
		log.Fatalf("Unable to create lite client: %v", err)
	}

	w, err := wallet.DefaultWalletFromSeed(SEED, client)
	if err != nil {
		log.Fatalf("Unable to create wallet: %v", err)
	}

	log.Printf("Wallet address: %v\n", w.GetAddress().ToRaw())
	err = w.Send(context.TODO(), wallet.SimpleTransfer{
		Amount:  ton.OneTON,
		Address: tongo.MustParseAccountID("EQBszTJahYw3lpP64ryqscKQaDGk4QpsO7RO6LYVvKHSINS0"),
		Comment: "hi! hope it will be enough for buying a yacht",
	})
	if err != nil {
		log.Fatalf("Unable to generate transfer message: %v", err)
	}
}
