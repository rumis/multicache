package utils

import (
	"fmt"
	"sync"
	"testing"
)

func TestSafeRand(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {

			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()

			r := SafeRand()
			n := r.Intn(10)
			if n < 0 || n >= 10 {
				t.Errorf("Intn() = %d, want [0, 100)", n)
			}
			fmt.Println(idx, ":", n)
			wg.Done()
		}(i)
	}

	wg.Wait()

}
