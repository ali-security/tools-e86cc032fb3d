// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/tools/internal/jsonrpc2"
)

func TestListenParsing(t *testing.T) {
	tests := []struct {
		input, wantNetwork, wantAddr string
	}{
		{"127.0.0.1:0", "tcp", "127.0.0.1:0"},
		{"unix;/tmp/sock", "unix", "/tmp/sock"},
		{"auto", "auto", ""},
		{"auto;foo", "auto", "foo"},
	}

	for _, test := range tests {
		gotNetwork, gotAddr := parseAddr(test.input)
		if gotNetwork != test.wantNetwork {
			t.Errorf("network = %q, want %q", gotNetwork, test.wantNetwork)
		}
		if gotAddr != test.wantAddr {
			t.Errorf("addr = %q, want %q", gotAddr, test.wantAddr)
		}
	}
}

// TestServeNoPortFlag checks that the -port flag, which bound the LSP server
// to all network interfaces (":<port>"), is no longer exposed (CVE-2026-42503).
func TestServeNoPortFlag(t *testing.T) {
	var check func(reflect.Type)
	check = func(typ reflect.Type) {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if name, ok := field.Tag.Lookup("flag"); ok && name == "port" {
				t.Errorf("%s.%s exposes the -port flag, which implicitly binds all network interfaces", typ.Name(), field.Name)
			}
			if field.Type.Kind() == reflect.Struct && field.PkgPath == "" {
				check(field.Type)
			}
		}
	}
	check(reflect.TypeOf(Serve{}))
	check(reflect.TypeOf(Application{}))
}

// TestServeRejectsImplicitAllInterfaces checks that -listen addresses with an
// empty host, which would implicitly bind all network interfaces, are rejected
// before any listener is created (CVE-2026-42503).
func TestServeRejectsImplicitAllInterfaces(t *testing.T) {
	app := New("gopls-test", "", nil, nil)
	for _, address := range []string{":0", "tcp;:0"} {
		// The idle timeout only matters if the address is (wrongly) accepted:
		// it makes the test fail promptly rather than serve forever.
		s := &Serve{Address: address, IdleTimeout: 50 * time.Millisecond, app: app}
		err := s.Run(context.Background())
		if err == nil {
			t.Errorf("Serve{Address: %q}.Run() succeeded, want error", address)
			continue
		}
		if !strings.Contains(err.Error(), "implicitly binds all network interfaces") {
			t.Errorf("Serve{Address: %q}.Run() = %v, want implicit all-interfaces error", address, err)
		}
	}
}

// TestServeAcceptsExplicitHost checks that -listen addresses with an explicit
// host are still served.
func TestServeAcceptsExplicitHost(t *testing.T) {
	app := New("gopls-test", "", nil, nil)
	for _, address := range []string{"localhost:0", "tcp;127.0.0.1:0"} {
		s := &Serve{Address: address, IdleTimeout: 50 * time.Millisecond, app: app}
		if err := s.Run(context.Background()); err != jsonrpc2.ErrIdleTimeout {
			t.Errorf("Serve{Address: %q}.Run() = %v, want %v", address, err, jsonrpc2.ErrIdleTimeout)
		}
	}
}
