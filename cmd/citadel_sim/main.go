package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("[*] Inicializando simulador avanzado de Deadlock: StrongBox / citadeld...")
	
	var mu sync.Mutex
	var wg sync.WaitGroup
	var timeouts int64
	workers := 4

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					locked := tryLockWithTimeout(&mu, 100*time.Millisecond)
					if !locked {
						atomic.AddInt64(&timeouts, 1)
						continue
					}

					// Simular carga pesada o bloqueo en hilo específico
					if id == 2 {
						time.Sleep(800 * time.Millisecond)
					} else {
						time.Sleep(50 * time.Millisecond)
					}
					mu.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("[+] Simulación finalizada. Total de timeouts / bloqueos SPI detectados: %d\n", atomic.LoadInt64(&timeouts))
}

func tryLockWithTimeout(mu *sync.Mutex, timeout time.Duration) bool {
	ch := make(chan struct{}, 1)
	go func() {
		mu.Lock()
		ch <- struct{}{}
	}()

	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return false
	}
}
