package dronutz

import (
	"time"

	"golang.org/x/net/context"
)

type KitchenService struct {
	transactionChan chan func(*Donuts)
	fryerChan       chan *Donut
}

func NewKitchenService(Config) *KitchenService {
	service := &KitchenService{
		transactionChan: make(chan func(*Donuts)),
		fryerChan:       make(chan *Donut, 3),
	}
	go service.run()
	return service
}

func (k *KitchenService) Add(ctx context.Context, donuts *Donuts) (*Empty, error) {
	for _, donut := range donuts.Donuts {
		k.fryerChan <- donut
	}
	return &Empty{}, nil
}

func (k *KitchenService) Check(ctx context.Context, _ *Empty) (*Donuts, error) {
	result := &Donuts{}
	k.transact(func(state *Donuts) {
		donuts := make([]*Donut, len(state.Donuts))
		copy(donuts, state.Donuts)
		result.Donuts = donuts
	})
	return result, nil
}

func (k *KitchenService) Remove(context.Context, *Donuts) (*Empty, error) {
	// TODO: maybe we should take some donuts out...
	return &Empty{}, nil
}

func (k *KitchenService) transact(exec func(*Donuts)) {
	complete := make(chan struct{})
	k.transactionChan <- func(d *Donuts) {
		exec(d)
		close(complete)
	}
	<-complete
}

func (k *KitchenService) run() {
	state := &Donuts{}

	for {
		select {
		case transaction := <-k.transactionChan:
			transaction(state)
		case donut := <-k.fryerChan:
			switch donut.Status {
			case StatusNew:
				time.Sleep(400 * time.Millisecond)
				donut.Status = StatusReceived
				state.Donuts = append(state.Donuts, donut)
				go func() {
					time.Sleep(1 * time.Second)
					k.fryerChan <- donut
				}()
			case StatusReceived:
				time.Sleep(200 * time.Millisecond)
				donut.Status = StatusCooking
				go func() {
					time.Sleep(2 * time.Second)
					k.fryerChan <- donut
				}()
			case StatusCooking:
				time.Sleep(time.Second)
				donut.Status = StatusReady
			}
		}
	}
}
