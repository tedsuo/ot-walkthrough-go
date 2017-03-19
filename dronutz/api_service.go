package dronutz

import (
	"html/template"
	"net/http"

	"golang.org/x/net/context"
)

type APIservice struct {
	cfg        Config
	cfgTmpl    *template.Template
	kitchen    KitchenClient
	FileServer http.Handler
}

func NewAPIService(cfg Config, kitchen KitchenClient) *APIservice {
	return &APIservice{
		cfg:        cfg,
		kitchen:    kitchen,
		FileServer: http.FileServer(http.Dir(cfg.PublicDir)),
	}
}

func (api *APIservice) ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", api.FileServer)
	mux.HandleFunc("/config.js", api.Config)
	mux.HandleFunc("/order", api.Order)
	mux.HandleFunc("/status", api.Status)
	return mux
}

func (a *APIservice) Config(res http.ResponseWriter, req *http.Request) {
	err := configJSTemplate.Execute(res, a.cfg)
	if err != nil {
		writeErrorResponse(res, err)
	}
}

func (a *APIservice) Order(res http.ResponseWriter, req *http.Request) {
	order := OrderRequest{}
	orderId := guid()
	ctx := req.Context()

	readJSON(req.Body, &order)

	donuts := order.ToDonuts(orderId)
	_, err := a.kitchen.Add(ctx, donuts)
	if err != nil {
		writeErrorResponse(res, err)
		return
	}

	statusRes, err := a.checkStatus(ctx, orderId)
	if err != nil {
		writeErrorResponse(res, err)
		return
	}

	writeJSON(res, statusRes)
}

func (a *APIservice) Status(res http.ResponseWriter, req *http.Request) {
	var statusReq StatusReq
	readJSON(req.Body, &statusReq)

	statusRes, err := a.checkStatus(req.Context(), statusReq.OrderId)
	if err != nil {
		writeErrorResponse(res, err)
		return
	}

	writeJSON(res, statusRes)
}

func (a *APIservice) checkStatus(ctx context.Context, orderId string) (StatusRes, error) {
	donuts, err := a.kitchen.Check(ctx, &Empty{})
	if err != nil {
		return StatusRes{}, err
	}

	return estimateOrderStatus(orderId, donuts), nil
}
