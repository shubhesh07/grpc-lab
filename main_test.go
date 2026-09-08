package main

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestParseMethods(t *testing.T) {
	d := `testing.TestService is a service:
service TestService {
  rpc EmptyCall ( .testing.Empty ) returns ( .testing.Empty );
  rpc FullDuplexCall ( stream .testing.Req ) returns ( stream .testing.Resp );
  rpc StreamingOutputCall ( .testing.Req ) returns ( stream .testing.Resp );
}`
	got := parseMethods("testing.TestService", d)
	want := []method{
		{Name: "testing.TestService.EmptyCall", Input: "testing.Empty", Output: "testing.Empty"},
		{Name: "testing.TestService.FullDuplexCall", Input: "testing.Req", Output: "testing.Resp", ClientStream: true, ServerStream: true},
		{Name: "testing.TestService.StreamingOutputCall", Input: "testing.Req", Output: "testing.Resp", ServerStream: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
}

func TestExtractTypeRefs(t *testing.T) {
	d := `message Cart {
  .shared.common.v1.Seller seller = 1;
  map<string, .cart.v1.Item> items = 2;
  repeated .google.protobuf.Any extras = 3;
  int32 qty = 4;
}`
	got := extractTypeRefs(d)
	want := []string{"shared.common.v1.Seller", "cart.v1.Item", "google.protobuf.Any"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseHeadersAndQuote(t *testing.T) {
	got := parseHeaders("x-tenant: tm\n# comment\n\nnot-a-header\nx-b: 1")
	want := []string{"-H", "x-tenant: tm", "-H", "x-b: 1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if q := shellQuote([]string{"grpcurl", "-d", `{"a":"it's"}`, "localhost:1"}); q != `grpcurl -d '{"a":"it'\''s"}' localhost:1` {
		t.Fatalf("bad quote: %s", q)
	}
}

func TestParseFields(t *testing.T) {
	d := `cart.v1.Cart is a message:
message Cart {
  string id = 1;
  .cart.v1.CartType cart_type = 2;
  map<string, .cart.v1.LineItem> items = 21;
  repeated .cart.v1.Cart child_carts = 22;
  .google.protobuf.Any seller_info = 54;
  map<string, .google.protobuf.Any> shared_contexts = 2;
}`
	got := parseFields(d)
	want := map[string]field{
		"id":             {Type: "string", Kind: "scalar"},
		"cartType":       {Type: "cart.v1.CartType", Kind: "msg"},
		"items":          {Type: "cart.v1.LineItem", Kind: "msg", Map: true},
		"childCarts":     {Type: "cart.v1.Cart", Kind: "msg", Repeated: true},
		"sellerInfo":     {Type: "google.protobuf.Any", Kind: "any"},
		"sharedContexts": {Type: "google.protobuf.Any", Kind: "any", Map: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
}

func TestCompileProto(t *testing.T) {
	if _, err := exec.LookPath("protoc"); err != nil {
		t.Skip("protoc not installed")
	}
	src := []byte(`syntax = "proto3"; package lab; import "google/protobuf/timestamp.proto";
message Ping { google.protobuf.Timestamp at = 1; }`)
	b, err := compileProto("ping", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 || !strings.Contains(string(b), "google/protobuf/timestamp.proto") {
		t.Fatalf("descriptor set missing imported file (%d bytes)", len(b))
	}
	if _, err := compileProto("bad", []byte("syntax = nope;")); err == nil {
		t.Fatal("expected protoc error")
	}
}

func TestSafeReqName(t *testing.T) {
	ok := []string{"a", "cart/basic", "v1.2_x-y", "a/b/c"}
	bad := []string{"", "../x", "a/../x", "a/b/c/d/e/f/g/h/i", "./x", "a/", "/a", "a b"}
	for _, n := range ok {
		if !safeReqName(n) {
			t.Errorf("%q should be allowed", n)
		}
	}
	for _, n := range bad {
		if safeReqName(n) {
			t.Errorf("%q should be rejected", n)
		}
	}
}
