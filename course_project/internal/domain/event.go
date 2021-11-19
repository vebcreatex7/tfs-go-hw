package domain

type Event struct {
	Event      string        `json:"event"`
	Feed       string        `json:"feed,omitempty"`
	ProductIds []interface{} `json:"product_ids,omitempty"`
	Version    int           `json:"version,omitempty"`
}
