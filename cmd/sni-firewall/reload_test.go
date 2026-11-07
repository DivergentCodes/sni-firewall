package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestReloadDomainListOnSignal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domains.txt")
	writeDomains := func(contents string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	writeDomains("before.example\n")
	if err := loadDomainList(path); err != nil {
		t.Fatal(err)
	}

	signals := make(chan os.Signal)
	results := make(chan error)
	go reloadDomainList(path, signals, results)
	defer close(signals)

	writeDomains("after.example\n")
	signals <- syscall.SIGHUP
	if err := <-results; err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if _, ok := isDomainInList("before.example"); ok {
		t.Error("before.example still matched after reload")
	}
	if _, ok := isDomainInList("after.example"); !ok {
		t.Error("after.example not matched after reload")
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	signals <- syscall.SIGHUP
	if err := <-results; err == nil {
		t.Fatal("reload of a missing file succeeded")
	}
	if _, ok := isDomainInList("after.example"); !ok {
		t.Error("a failed reload replaced the current list")
	}
}
