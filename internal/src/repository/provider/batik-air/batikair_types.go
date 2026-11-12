package batikair

import "time"

var batikAirTimeLayouts = []string{
	time.RFC3339,               // 2025-12-15T07:15:00+07:00
	"2006-01-02T15:04:05-0700", // 2025-12-15T07:15:00+0700
}

type BatikAirSearchResp struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Results []Flight `json:"results"`
}

type Flight struct {
	FlightNumber      string       `json:"flightNumber"`
	AirlineName       string       `json:"airlineName"`
	AirlineIATA       string       `json:"airlineIATA"`
	Origin            string       `json:"origin"`
	Destination       string       `json:"destination"`
	DepartureDateTime string       `json:"departureDateTime"`
	ArrivalDateTime   string       `json:"arrivalDateTime"`
	TravelTime        string       `json:"travelTime"`
	NumberOfStops     int          `json:"numberOfStops"`
	Connections       []Connection `json:"connections"`
	Fare              Fare         `json:"fare"`
	SeatsAvailable    int          `json:"seatsAvailable"`
	AircraftModel     string       `json:"aircraftModel"`
	BaggageInfo       string       `json:"baggageInfo"`
	OnboardServices   []string     `json:"onboardServices"`
}

type Fare struct {
	BasePrice    float64 `json:"basePrice"`
	Taxes        float64 `json:"taxes"`
	TotalPrice   float64 `json:"totalPrice"`
	CurrencyCode string  `json:"currencyCode"`
	Class        string  `json:"class"`
}

type Connection struct {
	StopAirPort  string `json:"stopAirport"`
	StopDuration string `json:"stopDuration"`
}
