package cxdbmemory

import (
	"fmt"
	"sync"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// MemoryAuctionOrderbook is the representation of a auction orderbook for SQL
type MemoryAuctionOrderbook struct {
	// orders is a map from auction id to a map of price -> list of orders
	orders  map[match.AuctionID]map[float64][]*match.AuctionOrderIDPair
	bookMtx *sync.Mutex

	// this pair
	pair *match.Pair
}

// CreateAuctionOrderbook creates a auction orderbook based on a pair
func CreateAuctionOrderbook(pair *match.Pair) (book match.AuctionOrderbook, err error) {
	// Set values for auction engine
	mo := &MemoryAuctionOrderbook{
		pair:    pair,
		orders:  make(map[match.AuctionID]map[float64][]*match.AuctionOrderIDPair),
		bookMtx: new(sync.Mutex),
	}
	// We can connect, now set return
	book = mo
	return
}

// UpdateBookExec takes in an order execution and updates the orderbook.
func (mo *MemoryAuctionOrderbook) UpdateBookExec(exec *match.OrderExecution) (err error) {
	mo.bookMtx.Lock()
	defer mo.bookMtx.Unlock()

	for aucID, priceMap := range mo.orders {
		for pr, pairList := range priceMap {
			for idx, pair := range pairList {
				if pair.OrderID == exec.OrderID {
					if exec.Filled {
						// remove
						last := len(pairList) - 1
						pairList[idx] = pairList[last]
						mo.orders[aucID][pr] = pairList[:last]
						if len(mo.orders[aucID][pr]) == 0 {
							delete(mo.orders[aucID], pr)
						}
					} else {
						pair.Order.AmountHave = exec.NewAmountHave
						pair.Order.AmountWant = exec.NewAmountWant
					}
					return
				}
			}
		}
	}
	err = fmt.Errorf("order not found")
	return
}

// UpdateBookCancel takes in an order cancellation and updates the orderbook.
func (mo *MemoryAuctionOrderbook) UpdateBookCancel(cancel *match.CancelledOrder) (err error) {
	mo.bookMtx.Lock()
	defer mo.bookMtx.Unlock()

	for aucID, priceMap := range mo.orders {
		for pr, pairList := range priceMap {
			for idx, pair := range pairList {
				if pair.OrderID == *cancel.OrderID {
					last := len(pairList) - 1
					pairList[idx] = pairList[last]
					mo.orders[aucID][pr] = pairList[:last]
					if len(mo.orders[aucID][pr]) == 0 {
						delete(mo.orders[aucID], pr)
					}
					return
				}
			}
		}
	}

	err = fmt.Errorf("order not found")
	return
}

// UpdateBookPlace takes in an order, ID, auction ID, and adds the order to the orderbook.
func (mo *MemoryAuctionOrderbook) UpdateBookPlace(auctionIDPair *match.AuctionOrderIDPair) (err error) {
	mo.bookMtx.Lock()
	defer mo.bookMtx.Unlock()

	aid := auctionIDPair.Order.AuctionID
	pr := auctionIDPair.Price

	if _, ok := mo.orders[aid]; !ok {
		mo.orders[aid] = make(map[float64][]*match.AuctionOrderIDPair)
	}
	mo.orders[aid][pr] = append(mo.orders[aid][pr], auctionIDPair)
	return
}

// GetOrder gets an order from an OrderID
func (mo *MemoryAuctionOrderbook) GetOrder(orderID *match.OrderID) (aucOrder *match.AuctionOrderIDPair, err error) {
	mo.bookMtx.Lock()
	defer mo.bookMtx.Unlock()

	for _, priceMap := range mo.orders {
		for _, pairList := range priceMap {
			for _, pair := range pairList {
				if pair.OrderID == *orderID {
					aucOrder = pair
					return
				}
			}
		}
	}
	err = fmt.Errorf("order not found")
	return
}

// CalculatePrice takes in a pair and returns the calculated price based on the orderbook.
func (mo *MemoryAuctionOrderbook) CalculatePrice(auctionID *match.AuctionID) (price float64, err error) {
	mo.bookMtx.Lock()
	defer mo.bookMtx.Unlock()

	orderMap, ok := mo.orders[*auctionID]
	if !ok {
		err = fmt.Errorf("auction not found")
		return
	}

	var maxSell float64
	var minBuy float64
	minBuy = 0
	firstBuy := true
	for pr, list := range orderMap {
		for _, pair := range list {
			if pair.Order.IsSellSide() {
				if pr > maxSell {
					maxSell = pr
				}
			} else if pair.Order.IsBuySide() {
				if firstBuy || pr < minBuy {
					minBuy = pr
					firstBuy = false
				}
			}
		}
	}
	price = (minBuy + maxSell) / 2
	return
}

// GetOrdersForPubkey gets orders for a specific pubkey.
func (mo *MemoryAuctionOrderbook) GetOrdersForPubkey(pubkey *koblitz.PublicKey) (orders map[float64][]*match.AuctionOrderIDPair, err error) {
	orders = make(map[float64][]*match.AuctionOrderIDPair)
	mo.bookMtx.Lock()
	defer mo.bookMtx.Unlock()

	var pk [33]byte
	copy(pk[:], pubkey.SerializeCompressed())

	for _, priceMap := range mo.orders {
		for pr, pairList := range priceMap {
			for _, pair := range pairList {
				if pair.Order.Pubkey == pk {
					orders[pr] = append(orders[pr], pair)
				}
			}
		}
	}
	return
}

// ViewAuctionOrderbook takes in a trading pair and returns the orderbook as a map
func (mo *MemoryAuctionOrderbook) ViewAuctionOrderBook() (book map[float64][]*match.AuctionOrderIDPair, err error) {
	book = make(map[float64][]*match.AuctionOrderIDPair)
	mo.bookMtx.Lock()
	defer mo.bookMtx.Unlock()

	for _, priceMap := range mo.orders {
		for pr, pairList := range priceMap {
			book[pr] = append(book[pr], pairList...)
		}
	}
	return
}

// CreateAuctionOrderbookMap creates a map of pair to auction engine, given a list of pairs.
func CreateAuctionOrderbookMap(pairList []*match.Pair) (aucMap map[match.Pair]match.AuctionOrderbook, err error) {

	aucMap = make(map[match.Pair]match.AuctionOrderbook)
	var curAucEng match.AuctionOrderbook
	for _, pair := range pairList {
		if curAucEng, err = CreateAuctionOrderbook(pair); err != nil {
			err = fmt.Errorf("Error creating single auction engine while creating auction engine map: %s", err)
			return
		}
		aucMap[*pair] = curAucEng
	}

	return
}
