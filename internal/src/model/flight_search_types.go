package model

import "time"

type SearchInput struct {
	Origin        string     `json:"origin"`
	Destination   string     `json:"destination"`
	DepartureDate time.Time  `json:"departureDate"`
	ReturnDate    *time.Time `json:"returnDate"`
	Passengers    int        `json:"passengers"`
	CabinClass    string     `json:"cabinClass"`
	Sort          string     `json:"sort"`
}

type SearchResult struct {
	TotalResults       int      `json:"total_results"`
	ProvidersQueried   int      `json:"providers_queried"`
	ProvidersSucceeded int      `json:"providers_succeeded"`
	ProvidersFailed    int      `json:"providers_failed"`
	SearchTimeMs       int64    `json:"search_time_ms"`
	Flights            []Flight `json:"flights"`
}

type SearchResponse struct {
	SearchCriteria SearchCriteria `json:"search_criteria"`
	Metadata       SearchMetadata `json:"metadata"`
	Flights        []Flight       `json:"flights"`
}

type SearchMetadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
}

type SearchCriteria struct {
	Origin        string  `json:"origin"`
	Destination   string  `json:"destination"`
	DepartureDate string  `json:"departure_date"`
	ReturnDate    *string `json:"return_date,omitempty"`
	Passengers    int     `json:"passengers"`
	CabinClass    string  `json:"cabin_class"`
}

type Flight struct {
	ID             string    `json:"id"`
	Provider       string    `json:"provider"`
	Airline        AirLine   `json:"airline"`
	FlightNumber   string    `json:"flight_number"`
	Departure      Departure `json:"departure"`
	Arrival        Arrival   `json:"arrival"`
	Duration       Duration  `json:"duration"`
	Stops          int       `json:"stops"`
	Price          Price     `json:"price"`
	AvailableSeats int64     `json:"available_seats"`
	CabinClass     string    `json:"cabin_class"`
	Aircraft       *Aircraft `json:"aircraft"`
	Amenities      []string  `json:"amenities"`
	Baggage        Baggage   `json:"baggage"`
}

type Aircraft struct {
	Model string `json:"model"`
}

type AirLine struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Departure struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	DateTime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
}

type Arrival struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	DateTime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
}

type Duration struct {
	TotalMinutes int64  `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Baggage struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}
