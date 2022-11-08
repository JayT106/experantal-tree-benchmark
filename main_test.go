package treebenchmark

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	mRand "math/rand"
	"sort"
	"testing"
	"time"

	"github.com/cosmos/iavl"
	verkle "github.com/gballet/go-verkle"
	db "github.com/tendermint/tm-db"
)

var fourtyKeyTest, _ = hex.DecodeString("4000000000000000000000000000000000000000000000000000000000000000")

func benchmarkCommitVerkelNLeaves(b *testing.B, n int) {
	type kv struct {
		k []byte
		v []byte
	}
	kvs := make([]kv, n)
	sortedKVs := make([]kv, n)

	for i := 0; i < n; i++ {
		key := make([]byte, 32)
		val := make([]byte, 32)
		rand.Read(key) // skipcq: GSC-G404
		rand.Read(val) // skipcq: GSC-G404
		kvs[i] = kv{k: key, v: val}
		sortedKVs[i] = kv{k: key, v: val}
	}

	// InsertOrder assumes keys are sorted
	sortKVs := func(src []kv) {
		sort.Slice(src, func(i, j int) bool { return bytes.Compare(src[i].k, src[j].k) < 0 })
	}
	sortKVs(sortedKVs)

	b.Run(fmt.Sprintf("insert/leaves/%d", n), func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			root := verkle.New()
			for _, el := range kvs {
				if err := root.Insert(el.k, el.v, nil); err != nil {
					b.Error(err)
				}
			}
			root.Commit()
		}
	})

	b.Run(fmt.Sprintf("insertOrdered/leaves/%d", n), func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			root := verkle.New()
			for _, el := range sortedKVs {
				if err := root.InsertOrdered(el.k, el.v, nil); err != nil {
					b.Fatal(err)
				}
			}
			root.Commit()
		}
	})
}

func benchmarkCommitIAVLNLeaves(b *testing.B, n int) {
	type kv struct {
		k []byte
		v []byte
	}
	kvs := make([]kv, n)
	sortedKVs := make([]kv, n)

	for i := 0; i < n; i++ {
		key := make([]byte, 32)
		val := make([]byte, 32)
		rand.Read(key) // skipcq: GSC-G404
		rand.Read(val) // skipcq: GSC-G404
		kvs[i] = kv{k: key, v: val}
		sortedKVs[i] = kv{k: key, v: val}
	}

	// InsertOrder assumes keys are sorted
	sortKVs := func(src []kv) {
		sort.Slice(src, func(i, j int) bool { return bytes.Compare(src[i].k, src[j].k) < 0 })
	}
	sortKVs(sortedKVs)

	b.Run(fmt.Sprintf("insert/leaves/%d", n), func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			mdb := db.NewMemDB()
			t, _ := iavl.NewMutableTree(mdb, 0, true)
			for _, el := range kvs {
				if _, err := t.Set(el.k, el.v); err != nil {
					b.Fatal(err)

				}
			}
			t.SaveVersion()
		}
	})

	b.Run(fmt.Sprintf("insertOrdered/leaves/%d", n), func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			root := verkle.New()
			for _, el := range sortedKVs {
				if err := root.InsertOrdered(el.k, el.v, nil); err != nil {
					b.Fatal(err)
				}
			}
			root.Commit()
		}
	})
}

func BenchmarkVerkleModifyLeaves(b *testing.B) {
	mRand.Seed(time.Now().UnixNano()) // skipcq: GO-S1033

	n := 200000
	toEdit := 10000
	val := []byte{0}
	keys := make([][]byte, n)
	root := verkle.New()
	for i := 0; i < n; i++ {
		key := make([]byte, 32)
		rand.Read(key) // skipcq: GSC-G404
		keys[i] = key
		root.Insert(key, val, nil)
	}
	root.Commit()

	b.ResetTimer()
	b.ReportAllocs()

	val = make([]byte, 4)
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint32(val, uint32(i))
		for j := 0; j < toEdit; j++ {
			// skipcq: GSC-G404
			k := keys[mRand.Intn(n)]
			if err := root.Insert(k, val, nil); err != nil {
				b.Error(err)
			}
		}
		root.Commit()
	}
}

func BenchmarkIAVLModifyLeaves(b *testing.B) {
	mRand.Seed(time.Now().UnixNano()) // skipcq: GO-S1033

	n := 200000
	toEdit := 10000
	val := []byte{0}
	keys := make([][]byte, n)

	mdb := db.NewMemDB()
	t, _ := iavl.NewMutableTree(mdb, 0, false)
	for i := 0; i < n; i++ {
		key := make([]byte, 32)
		rand.Read(key) // skipcq: GSC-G404
		keys[i] = key
		t.Set(key, val)
	}
	t.SaveVersion()

	b.ResetTimer()
	b.ReportAllocs()

	val = make([]byte, 4)
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint32(val, uint32(i))
		for j := 0; j < toEdit; j++ {
			// skipcq: GSC-G404
			k := keys[mRand.Intn(n)]
			if _, err := t.Set(k, val); err != nil {
				b.Error(err)
			}
		}
		t.SaveVersion()
	}
}

func BenchmarkCommitVerkleLeaves(b *testing.B) {
	benchmarkCommitVerkelNLeaves(b, 1000)
	benchmarkCommitVerkelNLeaves(b, 10000)
}

func BenchmarkCommitIAVLLeaves(b *testing.B) {
	benchmarkCommitIAVLNLeaves(b, 1000)
	benchmarkCommitIAVLNLeaves(b, 10000)
}

func BenchmarkCommitVerkleFullNode(b *testing.B) {
	nChildren := 256
	keys := make([][]byte, nChildren)
	for i := 0; i < nChildren; i++ {
		key := make([]byte, 32)
		key[0] = uint8(i)
		keys[i] = key
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		root := verkle.New()
		for _, k := range keys {
			if err := root.Insert(k, fourtyKeyTest, nil); err != nil {
				b.Fatal(err)
			}
		}
		root.Commit()
	}
}

func BenchmarkCommitIavlFullNode(b *testing.B) {
	nChildren := 256
	keys := make([][]byte, nChildren)
	for i := 0; i < nChildren; i++ {
		key := make([]byte, 32)
		key[0] = uint8(i)
		keys[i] = key
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mdb := db.NewMemDB()
		t, _ := iavl.NewMutableTree(mdb, 0, false)
		for _, k := range keys {
			if _, err := t.Set(k, fourtyKeyTest); err != nil {
				b.Fatal(err)

			}
		}
		t.SaveVersion()
	}
}
