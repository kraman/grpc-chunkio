package main

import (
	"context"
	"crypto/rand"
	"io/ioutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if _, err := os.Stat("infile"); err != nil {
		buf := make([]byte, 1024*1024*1000) // 1GB
		if _, err := rand.Read(buf); err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile("infile", buf, 0660); err != nil {
			panic(err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := listen(ctx, ":30001"); err != nil {
			panic(err)
		}
	}()

	os.Exit(m.Run())
}

func TestTransfer(t *testing.T) {
	if err := downloadFile("infile", "outfile"); err != nil {
		t.Fatal(err)
	}

	b1, err := ioutil.ReadFile("infile")
	if err != nil {
		t.Fatal(err)
	}
	b2, err := ioutil.ReadFile("outfile")

	if err != nil {
		t.Fatal(err)
	}

	for i := range b1 {
		if b2[i] != b1[i] {
			t.Fatal("transfered file not equal")
		}
	}
}

func BenchmarkTransfer1G(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := downloadFile("infile", "outfile"); err != nil {
			b.Fatal(err)
		}
	}
}
