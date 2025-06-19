package cxdbmemory

import (
	"testing"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"golang.org/x/crypto/sha3"

	"github.com/mit-dci/opencx/match"
)

func createTestOrder(t *testing.T, pair *match.Pair) (*match.AuctionOrderIDPair, *koblitz.PublicKey) {
	priv, err := koblitz.NewPrivateKey(koblitz.S256())
	if err != nil {
		t.Fatalf("key gen err: %v", err)
	}
	pub := priv.PubKey()

	order := &match.AuctionOrder{
		TradingPair: *pair,
		Side:        match.Buy,
		AmountHave:  1000,
		AmountWant:  2000,
		AuctionID:   match.AuctionID{0x01},
		Nonce:       [2]byte{0x01, 0x02},
	}
	copy(order.Pubkey[:], pub.SerializeCompressed())
	pr, _ := order.Price()
	id := sha3.Sum256(order.SerializeSignable())
	pairStruct := &match.AuctionOrderIDPair{OrderID: id, Order: order, Price: pr}
	return pairStruct, pub
}

func TestMemoryAuctionOrderbookPlaceGet(t *testing.T) {
	btc, _ := match.AssetFromCoinParam(&coinparam.RegressionNetParams)
	ltc, _ := match.AssetFromCoinParam(&coinparam.LiteRegNetParams)
	pair := &match.Pair{AssetWant: btc, AssetHave: ltc}

	bookInt, err := CreateAuctionOrderbook(pair)
	if err != nil {
		t.Fatalf("create orderbook: %v", err)
	}
	book := bookInt.(*MemoryAuctionOrderbook)

	orderPair, _ := createTestOrder(t, pair)
	if err = book.UpdateBookPlace(orderPair); err != nil {
		t.Fatalf("place err: %v", err)
	}

	got, err := book.GetOrder(&orderPair.OrderID)
	if err != nil {
		t.Fatalf("get order err: %v", err)
	}
	if got.OrderID != orderPair.OrderID {
		t.Errorf("order id mismatch")
	}
}

func TestMemoryAuctionOrderbookCancel(t *testing.T) {
	btc, _ := match.AssetFromCoinParam(&coinparam.RegressionNetParams)
	ltc, _ := match.AssetFromCoinParam(&coinparam.LiteRegNetParams)
	pair := &match.Pair{AssetWant: btc, AssetHave: ltc}

	bookInt, err := CreateAuctionOrderbook(pair)
	if err != nil {
		t.Fatalf("create orderbook: %v", err)
	}
	book := bookInt.(*MemoryAuctionOrderbook)

	orderPair, _ := createTestOrder(t, pair)
	if err = book.UpdateBookPlace(orderPair); err != nil {
		t.Fatalf("place err: %v", err)
	}

	cancel := &match.CancelledOrder{OrderID: &orderPair.OrderID}
	if err = book.UpdateBookCancel(cancel); err != nil {
		t.Fatalf("cancel err: %v", err)
	}

	if _, err = book.GetOrder(&orderPair.OrderID); err == nil {
		t.Errorf("expected error retrieving cancelled order")
	}
}

func TestMemoryAuctionOrderbookOrdersForPubkey(t *testing.T) {
	btc, _ := match.AssetFromCoinParam(&coinparam.RegressionNetParams)
	ltc, _ := match.AssetFromCoinParam(&coinparam.LiteRegNetParams)
	pair := &match.Pair{AssetWant: btc, AssetHave: ltc}

	bookInt, err := CreateAuctionOrderbook(pair)
	if err != nil {
		t.Fatalf("create orderbook: %v", err)
	}
	book := bookInt.(*MemoryAuctionOrderbook)

	orderPair, pub := createTestOrder(t, pair)
	if err = book.UpdateBookPlace(orderPair); err != nil {
		t.Fatalf("place err: %v", err)
	}

	orders, err := book.GetOrdersForPubkey(pub)
	if err != nil {
		t.Fatalf("get orders err: %v", err)
	}
	pr := orderPair.Price
	if list, ok := orders[pr]; !ok || len(list) != 1 {
		t.Errorf("order for pubkey not found")
	}
}
