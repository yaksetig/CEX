package cxdbmemory

import (
	"fmt"
	"sync"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/match"
)

type pendingDeposit struct {
	pubkey  [33]byte
	amount  uint64
	confirm uint64
}

// MemoryDepositStore is a simple in-memory implementation of cxdb.DepositStore
// used for development and testing without a SQL database.
type MemoryDepositStore struct {
	coin *coinparam.Params

	addrToPub map[string]*koblitz.PublicKey
	pubToAddr map[[33]byte]string
	pending   map[uint64][]pendingDeposit

	mtx *sync.Mutex
}

// CreateDepositStore creates a deposit store for a specific coin.
func CreateDepositStore(coin *coinparam.Params) (store cxdb.DepositStore, err error) {
	md := &MemoryDepositStore{
		coin:      coin,
		addrToPub: make(map[string]*koblitz.PublicKey),
		pubToAddr: make(map[[33]byte]string),
		pending:   make(map[uint64][]pendingDeposit),
		mtx:       new(sync.Mutex),
	}
	store = md
	return
}

// CreateDepositStoreMap creates a map of coin to deposit store for a list of coins.
func CreateDepositStoreMap(coinList []*coinparam.Params) (depositMap map[*coinparam.Params]cxdb.DepositStore, err error) {
	depositMap = make(map[*coinparam.Params]cxdb.DepositStore)
	var cur cxdb.DepositStore
	for _, coin := range coinList {
		if cur, err = CreateDepositStore(coin); err != nil {
			return
		}
		depositMap[coin] = cur
	}
	return
}

// RegisterUser associates a pubkey with a deposit address.
func (md *MemoryDepositStore) RegisterUser(pubkey *koblitz.PublicKey, address string) (err error) {
	md.mtx.Lock()
	defer md.mtx.Unlock()

	var pk [33]byte
	copy(pk[:], pubkey.SerializeCompressed())
	md.pubToAddr[pk] = address
	md.addrToPub[address] = pubkey
	return
}

// UpdateDeposits updates pending deposits and returns settlement executions for
// deposits that mature at the provided block height.
func (md *MemoryDepositStore) UpdateDeposits(deposits []match.Deposit, blockheight uint64) (depositExecs []*match.SettlementExecution, err error) {
	var asset match.Asset
	if asset, err = match.AssetFromCoinParam(md.coin); err != nil {
		err = fmt.Errorf("Error getting asset from coin param for UpdateDeposits: %s", err)
		return
	}

	md.mtx.Lock()
	defer md.mtx.Unlock()

	// record new deposits
	for _, dep := range deposits {
		exp := dep.BlockHeightReceived + dep.Confirmations
		var pd pendingDeposit
		copy(pd.pubkey[:], dep.Pubkey.SerializeCompressed())
		pd.amount = dep.Amount
		pd.confirm = exp
		md.pending[exp] = append(md.pending[exp], pd)
	}

	if list, ok := md.pending[blockheight]; ok {
		for _, pd := range list {
			exec := &match.SettlementExecution{
				Pubkey: pd.pubkey,
				Amount: pd.amount,
				Asset:  asset,
				Type:   match.Debit,
			}
			depositExecs = append(depositExecs, exec)
		}
		delete(md.pending, blockheight)
	}

	return
}

// GetDepositAddressMap returns a copy of the address to pubkey map.
func (md *MemoryDepositStore) GetDepositAddressMap() (depAddrMap map[string]*koblitz.PublicKey, err error) {
	md.mtx.Lock()
	defer md.mtx.Unlock()

	depAddrMap = make(map[string]*koblitz.PublicKey)
	for addr, pk := range md.addrToPub {
		depAddrMap[addr] = pk
	}
	return
}

// GetDepositAddress gets the deposit address for a pubkey.
func (md *MemoryDepositStore) GetDepositAddress(pubkey *koblitz.PublicKey) (addr string, err error) {
	md.mtx.Lock()
	defer md.mtx.Unlock()

	var pk [33]byte
	copy(pk[:], pubkey.SerializeCompressed())
	var ok bool
	if addr, ok = md.pubToAddr[pk]; !ok {
		err = fmt.Errorf("address not found for pubkey")
	}
	return
}
