package provisions

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"math/big"

	"github.com/mit-dci/zksigma"
)

// helper to set balance in the internal map
func setBalance(pk *ecdsa.PublicKey, amt uint64) {
	key := string(elliptic.Marshal(pk.Curve, pk.X, pk.Y))
	balanceMtx.Lock()
	balanceMap[key] = amt
	balanceMtx.Unlock()
}

func TestBalRetrieval(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("error generating key: %v", err)
	}

	setBalance(&priv.PublicKey, 12345)
	bal, err := bal(&priv.PublicKey)
	if err != nil {
		t.Fatalf("bal returned error: %v", err)
	}
	if bal != 12345 {
		t.Fatalf("expected 12345 got %d", bal)
	}
}

func TestCalculateAssetCommitment(t *testing.T) {
	curve := zksigma.TestCurve.C
	machine, err := NewAssetsProofMachine(curve)
	if err != nil {
		t.Fatalf("error creating machine: %v", err)
	}

	priv1, _ := ecdsa.GenerateKey(curve, rand.Reader)
	priv2, _ := ecdsa.GenerateKey(curve, rand.Reader)

	machine.pkAnonSetMutex.Lock()
	machine.PubKeyAnonSet[&priv1.PublicKey] = priv1
	machine.PubKeyAnonSet[&priv2.PublicKey] = nil
	machine.pkAnonSetMutex.Unlock()

	setBalance(&priv1.PublicKey, 1000)
	setBalance(&priv2.PublicKey, 2000)

	commit, err := machine.CalculateAssetCommitment()
	if err != nil {
		t.Fatalf("error calculating asset commitment: %v", err)
	}

	expected := big.NewInt(1000)
	if !zksigma.Open(zksigma.TestCurve, expected, machine.assetRand, *commit) {
		t.Fatalf("commitment did not open with expected value")
	}
}
