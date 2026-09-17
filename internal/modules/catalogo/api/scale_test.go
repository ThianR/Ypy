package api

import "testing"

func TestScaleCodeExactRuleWins(t *testing.T) {
	got, err := ParseScaleCode("1234567890", []ScaleRule{{Code: "1234567890", Prefix: "12", WeightDigits: 3, PriceDigits: 5, Priority: 1}})
	if err != nil || got.ItemCode != "1234567890" {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}
func TestScaleCodeRejectsZeroPrice(t *testing.T) {
	if _, err := ParseScaleCode("9912300000", []ScaleRule{{Code: "MEAT", Prefix: "99", WeightDigits: 3, PriceDigits: 5}}); err != ErrInvalidScaleValue {
		t.Fatal("zero price accepted")
	}
}

func TestScaleCodePreservesFixedWidthDigits(t *testing.T) {
	got, err := ParseScaleCode("9912301234", []ScaleRule{{Code: "MEAT", Prefix: "99", WeightDigits: 3, PriceDigits: 5}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Quantity != "123" || got.Price != "01234" {
		t.Fatalf("expected fixed-width values, got %#v", got)
	}
}

func TestScaleCodeIgnoresMalformedRules(t *testing.T) {
	if _, err := ParseScaleCode("123456", []ScaleRule{{Code: "BAD", Prefix: "12", WeightDigits: -1, PriceDigits: 5}}); err != ErrInvalidScaleValue {
		t.Fatalf("expected malformed rule rejection, got %v", err)
	}
}

func TestParseV4ScaleCodeUsesConfiguredPositions(t *testing.T) {
	rule := V4ScaleRule{Prefix: "99", Length: 11, ProductStart: 3, ProductLength: 3, ValueStart: 6, ValueLength: 5, Decimals: 2, Content: "PRECIO"}
	got, err := ParseV4ScaleCode("99123001234", []V4ScaleRule{rule})
	if err != nil || got.ItemCode != "123" || got.Price != "00123" {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}

func TestParseV4ScaleCodeRejectsSignedFields(t *testing.T) {
	rule := V4ScaleRule{Prefix: "99", Length: 6, ProductStart: 3, ProductLength: 2, ValueStart: 5, ValueLength: 2, Decimals: 2, Content: "PRECIO"}
	if _, err := ParseV4ScaleCode("99+1+2", []V4ScaleRule{rule}); err != ErrInvalidScaleValue {
		t.Fatalf("signed fixed-width field accepted: %v", err)
	}
}
