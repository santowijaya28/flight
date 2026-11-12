package model

import "time"

type SearchInput struct {
	Origin        string     `json:"origin"`
	Destination   string     `json:"destination"`
	DepartureDate time.Time  `json:"departureDate"`
	ReturnDate    *time.Time `json:"returnDate"`
	Passengers    int        `json:"passengers"`
	CabinClass    string     `json:"cabinClass"`
}

type SearchResult struct {
	Flights []Flight `json:"flights"`
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
