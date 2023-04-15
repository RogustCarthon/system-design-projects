package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type BlockingQueue[T any] struct {
	Objects []T
	size    int
	m       *sync.Mutex
	c       *sync.Cond
}

func (bq *BlockingQueue[T]) PutBack(item T) error {
	bq.m.Lock()
	defer bq.m.Unlock()
	if len(bq.Objects) < bq.size {
		bq.Objects = append(bq.Objects, item)
		bq.c.Signal()
		return nil
	} else {
		return errors.New("queue is full")
	}
}

func (bq *BlockingQueue[T]) Pop() T {
	bq.m.Lock()
	defer bq.m.Unlock()
	if len(bq.Objects) == 0 {
		bq.c.Wait()
	}
	item := bq.Objects[0]
	bq.Objects = bq.Objects[1:]
	return item
}

func NewBlockingQueue[T any](size int) *BlockingQueue[T] {
	bq := &BlockingQueue[T]{
		Objects: make([]T, 0, size),
		size:    size,
		m:       &sync.Mutex{},
	}
	bq.c = sync.NewCond(bq.m)
	return bq
}

type Dummy struct {
	v int
}

func getAndReturn(routineId int, q *BlockingQueue[*Dummy], wg *sync.WaitGroup) {
	d := q.Pop()
	fmt.Printf("routine %v: got %v\n", routineId, d)
	time.Sleep(2 * time.Second)
	fmt.Printf("routine %v: putting back %v\n", routineId, d)
	q.PutBack(d)
	wg.Done()
}

func main() {
	q := NewBlockingQueue[*Dummy](5)
	for i := 0; i < 5; i++ {
		q.PutBack(&Dummy{i})
	}
	wg := &sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go getAndReturn(i, q, wg)
		time.Sleep(200 * time.Millisecond)
	}
	wg.Wait()
}
