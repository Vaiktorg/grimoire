package tests

import (
	"fmt"
	"github.com/vaiktorg/grimoire/names"
	"github.com/vaiktorg/grimoire/store"
	"github.com/vaiktorg/grimoire/uid"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestShardCache(t *testing.T) {
	t.Run("Get/Set", func(t *testing.T) {
		sc := store.NewShardCache[string](100)

		for _, name := range names.Names {
			sc.Set(name, string(uid.New()))
		}

		count := 0
		for _, name := range names.Names {
			if _, ok := sc.Get(name); ok {
				count++
			}
		}

		if count < 100 {
			t.Error("did not store all 100 items")
		}
	})
}

func BenchmarkShardCache(b *testing.B) {
	b.Run("Get/Set", func(b *testing.B) {
		cache := store.NewShardCache[string](b.N)

		for i := 0; i < b.N; i++ {
			cache.Set(strconv.Itoa(i), string(uid.New()))
		}

		for i := 0; i < b.N; i++ {
			fmt.Println(cache.Get(strconv.Itoa(i)))
		}

		for i := 0; i < b.N; i++ {
			cache.Delete(strconv.Itoa(i))
		}
	})
}
