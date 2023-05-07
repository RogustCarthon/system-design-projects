package distributed_kv_store

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInsert(t *testing.T) {
	s := Init()
	ctx := context.Background()

	keys := []string{}

	for i := 1; i < 6; i++ {
		for j := 0; j < 100; j++ {
			k := uuid.NewString()
			if err := s.Put(ctx, KVInput{
				Key:   k,
				Value: uuid.NewString(),
				TTL:   int64(i * 5),
			}); err != nil {
				t.Error(err)
			}
			keys = append(keys, k)
		}
	}

	fmt.Println(keys)
}

func TestKVStore(t *testing.T) {
	s := Init()
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	go s.Cleanup(ctx, wg)

	for i := 1; i < 6; i++ {
		for j := 0; j < 100; j++ {
			s.Put(ctx, KVInput{
				Key:   uuid.NewString(),
				Value: uuid.NewString(),
				TTL:   int64(i * 5),
			})
		}
	}

	a := time.After(35 * time.Second)
	for range time.NewTicker(time.Second).C {
		select {
		case <-a:
			break
		default:
			fmt.Println(s.GetCount(ctx))
		}
	}
	cancel()
	wg.Wait()
}
