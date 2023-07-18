package merge_two_asc_chan

import (
	"context"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestMergeTwoAscChan(t *testing.T) {
	testCases := []struct {
		name   string
		c1, c2 []int
		want   []int
	}{
		{
			name: "empty",
			c1:   []int{},
			c2:   []int{},
			want: []int{},
		},
		{
			name: "one on the first channel",
			c1:   []int{1},
			c2:   []int{},
			want: []int{1},
		},
		{
			name: "one on the second channel",
			c1:   []int{},
			c2:   []int{1},
			want: []int{1},
		},
		{
			name: "only on the first channel",
			c1:   []int{1, 2, 3, 4, 5, 6, 7},
			c2:   []int{},
			want: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name: "only on the second channel",
			c1:   []int{},
			c2:   []int{1, 2, 3, 4, 5, 6, 7},
			want: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name: "the same input",
			c1:   []int{1, 2, 3, 4, 5, 6, 7},
			c2:   []int{1, 2, 3, 4, 5, 6, 7},
			want: []int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7},
		},
		{
			name: "a lot on the first channel, one on the second channel",
			c1:   []int{2, 3, 4, 5, 6, 7},
			c2:   []int{1},
			want: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name: "a lot on the second channel, one on the first channel",
			c1:   []int{1},
			c2:   []int{2, 3, 4, 5, 6, 7},
			want: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name: "a lot on the first channel, one on the second channel #2",
			c1:   []int{1, 2, 3, 4, 5, 6},
			c2:   []int{7},
			want: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name: "a lot on the second channel, one on the first channel #2",
			c1:   []int{7},
			c2:   []int{1, 2, 3, 4, 5, 6},
			want: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name: "example 1",
			c1:   []int{1, 2, 6, 7, 8, 8, 11, 15},
			c2:   []int{3, 4, 7, 9, 10, 13},
			want: []int{1, 2, 3, 4, 6, 7, 7, 8, 8, 9, 10, 11, 13, 15},
		},
	}
	pour := func(dst chan int, src []int) {
		defer close(dst)
		for _, v := range src {
			dst <- v
		}
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancel()

			c1, c2 := make(chan int, 1), make(chan int, 1)
			go pour(c1, test.c1)
			go pour(c2, test.c2)

			gotChannel := MergeTwoAscChan(c1, c2)
			got := make([]int, 0, len(test.c1)+len(test.c2))
		loop:
			for {
				select {
				case <-ctx.Done():
					t.Error(ctx.Err())
					return
				case v, ok := <-gotChannel:
					if !ok {
						break loop
					}
					got = append(got, v)
				}
			}
			assert.Equal(t, test.want, got)
		})
	}
}

func TestMergeTwoAscChanRandom(t *testing.T) {
	const size = 1000
	const initial = -100000
	const maxStep = 1000
	want := make([]int, 0, size*2)
	var wantMx sync.Mutex
	var wantDelay time.Duration
	var wantDelayMx sync.Mutex

	pourRandom := func(dst chan<- int, wg *sync.WaitGroup) {
		defer wg.Done()
		defer close(dst)
		last := initial
		for i := 0; i < size; i++ {
			value := last + rand.Intn(maxStep)
			last = value
			if rand.Intn(4) == 0 {
				delay := time.Microsecond * time.Duration(rand.Intn(1000))
				wantDelayMx.Lock()
				wantDelay += delay
				wantDelayMx.Unlock()
				time.Sleep(delay)
			}
			wantMx.Lock()
			want = append(want, value)
			wantMx.Unlock()
			dst <- value
		}
	}

	c1, c2 := make(chan int, 1), make(chan int, 1)
	var wg sync.WaitGroup
	wg.Add(2)

	startTime := time.Now()
	go pourRandom(c1, &wg)
	go pourRandom(c2, &wg)
	gotChannel := MergeTwoAscChan(c1, c2)
	got := make([]int, 0, size*2)
	for value := range gotChannel {
		got = append(got, value)
	}
	duration := time.Since(startTime)

	wg.Wait()
	sort.Slice(want, func(i, j int) bool {
		return want[i] < want[j]
	})
	t.Logf("min: %d, max: %d, duration: %s, want duration: %s", want[0], want[len(want)-1], duration, wantDelay)

	assert.Less(t, duration, wantDelay)

	assert.Equal(t, want, got)
}
