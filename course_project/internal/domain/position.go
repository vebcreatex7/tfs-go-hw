package domain

type position struct {
	Side              string  `json:"side,omitempty"`
	Symbol            string  `json:"symbol,omitempty"`
	Price             float64 `json:"price,omitempty"`
	FillTime          string  `json:"fillTime,omitempty"`
	Size              int     `json:"size,omitempty"`
	UnrealizedFunding float64 `json:"unrealizedFunding,omitempty"`
}

type OpenPositions struct {
	Result        string     `json:"result,omitempty"`
	OpenPositions []position `json:"openPositions,omitempty"`
	ServerTime    string     `json:"serverTime,omitempty"`
	Error         string     `json:"error,omitempty"`
}
