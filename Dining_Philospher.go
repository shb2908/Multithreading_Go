/* Problem Statement

1. There should be 5 philosophers sharing chopsticks, with one chopstick between each adjacent pair of philosophers.

2. Each philosopher should eat only 3 times

3. The philosophers pick up the chopsticks in any order, not lowest-numbered first

4. In order to eat, a philosopher must get permission from a host which executes in its own goroutine.

5. The host allows no more than 2 philosophers to eat concurrently.
Each philosopher is numbered, 1 through 5.

6. When a philosopher starts eating (after it has obtained necessary locks) it prints “starting to eat <number>” on a line by itself, where <number> is the number of the philosopher.

7. When a philosopher finishes eating (before it has released its locks) it prints “finishing eating <number>” on a line by itself, where <number> is the number of the philosopher.

*/

package main

import (
	"fmt"
	"sync"
)

// chopsticks
type CSticks = sync.Mutex

// philosophers
type Philo struct {
	left, right *CSticks
	num         int
}

// channel int with buffer to keep track of total ongoing dine.
var c chan int

// execute eating :D
func eat(p *Philo) {

	//waiting for channel to free up
	select {

	//let the philo. eat when the channel is available.
	case c <- 1:
		p.left.Lock()
		p.right.Lock()

		fmt.Println("starting to eat ", p.num)
		fmt.Println("finishing eating ", p.num)

		p.right.Unlock()
		p.left.Unlock()
	}

	// free up the channel
	<-c
	//release waitgroup
	wg.Done()

}

// waitgroup
var wg sync.WaitGroup

// maintain a hosting facility.
func Host(p []*Philo) {

	defer wg.Done()

	for i := 0; i < 5; i++ {
		go eat(p[i])
	}

}

func main() {

	//making the chopstics
	css := make([]*CSticks, 5)

	for idx := 0; idx < 5; idx++ {
		css[idx] = new(CSticks)
	}

	philo := make([]*Philo, 5)
	for i := 0; i < 5; i++ {
		philo[i] = &Philo{css[i], css[(i+1)%5], i + 1}
	}

	c = make(chan int, 2)

	//executing the Host.
	wg.Add(18)
	go Host(philo)
	go Host(philo)
	go Host(philo)
	wg.Wait()

}
