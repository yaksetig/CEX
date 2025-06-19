package cxdbsql

import (
	"strings"
	"testing"
)

func TestSchemasUseIntegers(t *testing.T) {
	if strings.Contains(auctionOrderbookSchema, "DOUBLE") || strings.Contains(auctionEngineSchema, "DOUBLE") ||
		strings.Contains(limitOrderbookSchema, "DOUBLE") || strings.Contains(limitEngineSchema, "DOUBLE") {
		t.Errorf("schemas should not use floating point price columns")
	}
}
