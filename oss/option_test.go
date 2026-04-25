package oss

import (
	"testing"
	"time"
)

func TestApplyPutOptions(t *testing.T) {
	options := ApplyPutOptions(
		WithContentType("text/plain"),
		WithMetadata(map[string]string{"author": "gox"}),
	)

	if options.ContentType != "text/plain" {
		t.Fatalf("ContentType = %q, want %q", options.ContentType, "text/plain")
	}
	if options.Metadata["author"] != "gox" {
		t.Fatalf("Metadata[author] = %q, want %q", options.Metadata["author"], "gox")
	}
}

func TestApplyGetOptions(t *testing.T) {
	options := ApplyGetOptions()
	if options.RangeStart != -1 {
		t.Fatalf("RangeStart = %d, want -1", options.RangeStart)
	}

	options = ApplyGetOptions(WithRange(1, 10))
	if options.RangeStart != 1 || options.RangeEnd != 10 {
		t.Fatalf("range = (%d, %d), want (1, 10)", options.RangeStart, options.RangeEnd)
	}
}

func TestApplyListOptions(t *testing.T) {
	options := ApplyListOptions(
		WithPrefix("images/"),
		WithDelimiter("/"),
		WithLimit(10),
		WithToken("next"),
	)

	if options.Prefix != "images/" || options.Delimiter != "/" || options.Limit != 10 || options.Token != "next" {
		t.Fatalf("ListOptions = %+v", options)
	}
}

func TestApplySignOptions(t *testing.T) {
	options := ApplySignOptions()
	if options.Method != MethodGet {
		t.Fatalf("Method = %q, want %q", options.Method, MethodGet)
	}

	options = ApplySignOptions(
		WithMethod(MethodPut),
		WithExpires(time.Hour),
		WithSignContentType("image/png"),
	)
	if options.Method != MethodPut || options.Expires != time.Hour || options.ContentType != "image/png" {
		t.Fatalf("SignOptions = %+v", options)
	}
}
