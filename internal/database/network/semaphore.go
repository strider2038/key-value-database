package network

type Semaphore struct {
	tickets chan struct{}
}

func NewSemaphore(ticketsCount int) *Semaphore {
	return &Semaphore{tickets: make(chan struct{}, ticketsCount)}
}

func (s *Semaphore) Acquire() {
	s.tickets <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.tickets
}
