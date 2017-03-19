package dronutz

const (
	StatusNew      = "new"
	StatusReceived = "received"
	StatusCooking  = "cooking"
	StatusReady    = "ready"
)

type StatusReq struct {
	OrderId string `json:"order_id"`
}

type StatusRes struct {
	OrderId               string `json:"order_id"`
	EstimatedDeliveryTime int    `json:"estimated_delivery_time"`
	State                 string `json:"state"`
}

type DonutRequest struct {
	Flavor   string `json:"flavor"`
	Quantity int    `json:"quantity"`
}

type OrderRequest struct {
	Donuts []DonutRequest `json:"donuts"`
}

func (o *OrderRequest) ToDonuts(orderId string) *Donuts {
	donuts := &Donuts{}
	for _, donutReq := range o.Donuts {
		for i := 0; i < donutReq.Quantity; i++ {
			donuts.Donuts = append(donuts.Donuts, NewDonut(orderId, donutReq.Flavor))
		}
	}
	return donuts
}

func NewDonut(orderId string, flavor string) *Donut {
	return &Donut{
		OrderId: orderId,
		DonutId: guid(),
		Flavor:  flavor,
		Status:  StatusNew,
	}
}

func estimateOrderStatus(orderId string, donuts *Donuts) StatusRes {
	estState := StatusReady
	estTime := 0
	filteredDonuts := filterByOrderId(orderId, donuts)
	for _, donut := range filteredDonuts.Donuts {
		switch donut.Status {
		case StatusNew:
			estTime += 3
			estState = StatusNew
		case StatusReceived:
			estTime += 2
			estState = StatusReceived
		case StatusCooking:
			estTime++
			if estState != StatusReceived {
				estState = StatusCooking
			}
		}
	}
	return StatusRes{
		OrderId:               orderId,
		EstimatedDeliveryTime: estTime,
		State: estState,
	}
}

func filterByOrderId(orderId string, input *Donuts) *Donuts {
	result := &Donuts{}
	for _, donut := range input.Donuts {
		if donut.OrderId == orderId {
			result.Donuts = append(result.Donuts, donut)
		}
	}
	return result
}
