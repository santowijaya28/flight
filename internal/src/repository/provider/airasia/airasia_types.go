package airasia

type AirAsiaSearchResponse struct {
	Status  string   `json:"status"`
	Flights []Flight `json:"flights"`
}

type Flight struct {
	FlightCode    string  `json:"flight_code"`
	Airline       string  `json:"airline"`
	FromAirport   string  `json:"from_airport"`
	ToAirport     string  `json:"to_airport"`
	DepartTime    string  `json:"depart_time"`
	ArriveTime    string  `json:"arrive_time"`
	DurationHours float64 `json:"duration_hours"`
	DirectFlight  bool    `json:"direct_flight"`
	PriceIDR      float64 `json:"price_idr"`
	Seats         int64   `json:"seats"`
	CabinClass    string  `json:"cabin_class"`
	BaggageNotes  string  `json:"baggage_note"`
	Stops         []Stops `json:"stops"`
}

type Stops struct {
	Airport         string `json:"airport"`
	WaitTimeMinutes int64  `json:"wait_time_minutes"`
}

type Baggage struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}
