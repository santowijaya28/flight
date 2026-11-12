package garudaindonesia

type GarudaIndonesiaSearchResp struct {
	Status  string   `json:"status"`
	Flights []Flight `json:"flights"`
}

type Flight struct {
	FlightID        string     `json:"flight_id"`
	Airline         string     `json:"airline"`
	AirlineCode     string     `json:"airline_code"`
	Departure       Departure  `json:"departure"`
	Arrival         Arrival    `json:"arrival"`
	DurationMinutes int        `json:"duration_minutes"`
	Stops           int        `json:"stops"`
	Aircraft        string     `json:"aircraft"`
	Price           Price      `json:"price"`
	Segments        []Segmennt `json:"segments"`
	AvailableSeats  int        `json:"available_seats"`
	FareClass       string     `json:"fare_class"`
	Baggage         Baggage    `json:"baggage"`
	Amenities       []string   `json:"amenities"`
}

type Departure struct {
	Airport  string `json:"airport"`
	City     string `json:"city"`
	Time     string `json:"time"`
	Terminal string `json:"terminal"`
}

type Arrival struct {
	Airport  string `json:"airport"`
	City     string `json:"city"`
	Time     string `json:"time"`
	Terminal string `json:"terminal"`
}

type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Baggage struct {
	CarryOn int `json:"carry_on"`
	Checked int `json:"checked"`
}

type Segmennt struct {
	FlightNumber    string           `json:"flight_number"`
	Departure       SegmentDeparture `json:"departure"`
	Arrival         SegmentArrival   `json:"arrival"`
	DurationMinutes int              `json:"duration_minutes"`
	LayOverMinutes  int              `json:"layover_minutes"`
}

type SegmentDeparture struct {
	Airport string `json:"airport"`
	Time    string `json:"time"`
}

type SegmentArrival struct {
	Airport string `json:"airport"`
	Time    string `json:"time"`
}
