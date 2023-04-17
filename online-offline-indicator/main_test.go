package online_offline_indicator_test

import (
	"fmt"
	"online_offline_indicator"
	"online_offline_indicator/service"
	"sync"
	"testing"
	"time"
)

func client(
	t *testing.T,
	userId string,
	onlineFor time.Duration,
	svc *service.Service,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	goOffline := time.After(onlineFor)
	for range time.NewTicker(500 * time.Millisecond).C {
		select {
		case <-goOffline:
			t.Log(userId, "offline")
			return
		default:
			if err := svc.Heartbeat(userId); err != nil {
				t.Error(err)
			}
		}
	}
}

func TestBasic(t *testing.T) {
	svc := online_offline_indicator.GetService()

	clientCount := 10
	clientWg := &sync.WaitGroup{}
	clientWg.Add(clientCount)
	for i := 0; i < clientCount; i++ {
		go client(
			t, fmt.Sprintf("user_%d", i),
			2*time.Duration(i)*time.Second, svc,
			clientWg,
		)
	}
	checkerWg := &sync.WaitGroup{}
	checkerWg.Add(1)
	go func() {
		defer checkerWg.Done()
		for range time.NewTicker(time.Second).C {
			cnt, _ := svc.GetOnlineCount()
			t.Log("cnt", cnt)
			if cnt == 0 {
				return
			}
		}
	}()
	clientWg.Wait()
	checkerWg.Wait()
}
