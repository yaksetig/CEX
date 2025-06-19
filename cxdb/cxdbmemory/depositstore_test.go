package cxdbmemory

import (
	"testing"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/match"
)

func TestMemoryDepositStoreRegisterAndRetrieve(t *testing.T) {
	store, err := CreateDepositStore(&coinparam.BitcoinParams)
	if err != nil {
		t.Fatalf("create store err: %v", err)
	}

	priv, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), []byte{1})
	pub := priv.PubKey()
	addr := "addr1"

	if err = store.RegisterUser(pub, addr); err != nil {
		t.Fatalf("register user: %v", err)
	}

	if got, err := store.GetDepositAddress(pub); err != nil || got != addr {
		t.Fatalf("get deposit address mismatch: %v %s", err, got)
	}

	m, err := store.GetDepositAddressMap()
	if err != nil {
		t.Fatalf("get deposit map: %v", err)
	}
	if m[addr] == nil || !m[addr].IsEqual(pub) {
		t.Fatalf("map did not contain pubkey")
	}
}

func TestMemoryDepositStoreUpdateDeposits(t *testing.T) {
	storeIface, _ := CreateDepositStore(&coinparam.BitcoinParams)
	store := storeIface.(*MemoryDepositStore)

	priv, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), []byte{2})
	pub := priv.PubKey()
	addr := "addr2"
	_ = store.RegisterUser(pub, addr)

	dep := match.Deposit{
		Pubkey:              pub,
		Address:             addr,
		Amount:              100,
		Txid:                "tx",
		CoinType:            &coinparam.BitcoinParams,
		BlockHeightReceived: 5,
		Confirmations:       2,
	}

	execs, err := store.UpdateDeposits([]match.Deposit{dep}, 5)
	if err != nil {
		t.Fatalf("update deposits: %v", err)
	}
	if len(execs) != 0 {
		t.Fatalf("expected no execs yet")
	}

	execs, err = store.UpdateDeposits(nil, 7)
	if err != nil {
		t.Fatalf("second update: %v", err)
	}
	if len(execs) != 1 {
		t.Fatalf("expected execs on confirmation")
	}
	if execs[0].Amount != 100 {
		t.Fatalf("incorrect exec amount")
	}
	var exp [33]byte
	copy(exp[:], pub.SerializeCompressed())
	if execs[0].Pubkey != exp {
		t.Fatalf("incorrect exec pubkey")
	}
}
