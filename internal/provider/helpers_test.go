package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestParseImportID_valid(t *testing.T) {
	t.Parallel()
	resp := &resource.ImportStateResponse{}
	id, ok := parseImportID("42", resp)
	if !ok {
		t.Fatal("expected ok=true for valid numeric ID")
	}
	if id != 42 {
		t.Fatalf("expected id=42, got %d", id)
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func TestParseImportID_zero(t *testing.T) {
	t.Parallel()
	resp := &resource.ImportStateResponse{}
	id, ok := parseImportID("0", resp)
	if !ok {
		t.Fatal("expected ok=true for zero")
	}
	if id != 0 {
		t.Fatalf("expected id=0, got %d", id)
	}
}

func TestParseImportID_negative(t *testing.T) {
	t.Parallel()
	resp := &resource.ImportStateResponse{}
	id, ok := parseImportID("-5", resp)
	if !ok {
		t.Fatal("expected ok=true for negative integer")
	}
	if id != -5 {
		t.Fatalf("expected id=-5, got %d", id)
	}
}

func TestParseImportID_nonNumeric(t *testing.T) {
	t.Parallel()
	resp := &resource.ImportStateResponse{}
	_, ok := parseImportID("not-a-number", resp)
	if ok {
		t.Fatal("expected ok=false for non-numeric string")
	}
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics error for non-numeric string")
	}
}

func TestParseImportID_empty(t *testing.T) {
	t.Parallel()
	resp := &resource.ImportStateResponse{}
	_, ok := parseImportID("", resp)
	if ok {
		t.Fatal("expected ok=false for empty string")
	}
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics error for empty string")
	}
}

func TestParseImportID_float(t *testing.T) {
	t.Parallel()
	resp := &resource.ImportStateResponse{}
	_, ok := parseImportID("3.14", resp)
	if ok {
		t.Fatal("expected ok=false for float string")
	}
}
