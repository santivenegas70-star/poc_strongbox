package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("[*] Simulando Binder death y deadlock de SPI en citadeld / keystore2...")
	
	var mu sync.Mutex
	var wg sync.WaitGroup
	var timeouts int64
	var activeDeadlock int32

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if atomic.LoadInt32(&activeDeadlock) == 1 {
						time.Sleep(100 * time.Millisecond)
						atomic.AddInt64(&timeouts, 1)
						continue
					}

					locked := tryLockWithTimeout(&mu, 150*time.Millisecond)
					if !locked {
						atomic.AddInt64(&timeouts, 1)
						continue
					}

					// Simular que el proceso cliente muere abruptamente a mitad de operacion SPI
					if id == 3 {
						fmt.Printf("[!] ¡ALERTA: Hilo %d (Client) murio abruptamente con el bus SPI bloqueado!\n", id)
						atomic.StoreInt32(&activeDeadlock, 1)
						// El mutex se queda atrapado (no se hace unlock), simulando el freeze de citadeld
						return 
					}

					time.Sleep(60 * time.Millisecond)
					mu.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("[+] Simulación de colapso finalizada. Total de bloqueos / timeouts en keystore2: %d\n", atomic.LoadInt64(&timeouts))
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
