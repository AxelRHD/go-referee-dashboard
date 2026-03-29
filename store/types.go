package store

// Stammdaten

type League struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	Sorter    int    `json:"sorter"`
	Remarks   string `json:"remarks,omitempty"`
}

type Team struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	State    string `json:"state"`
	IsActive bool   `json:"is_active"`
	Remarks  string `json:"remarks,omitempty"`
}

type Venue struct {
	ID        string  `json:"id"`
	City      string  `json:"city"`
	ShortName string  `json:"short_name"`
	Stadium   string  `json:"stadium,omitempty"`
	Lat       float64 `json:"lat,omitempty"`
	Lon       float64 `json:"lon,omitempty"`
}

type Position struct {
	Position string `json:"position"`
	Long     string `json:"long"`
	Sorter   int    `json:"sorter"`
}

// Eingebettete Referenzen in Game-Dokumenten

type TeamRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type VenueRef struct {
	ID        string  `json:"id"`
	City      string  `json:"city"`
	ShortName string  `json:"short_name"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
}

type LeagueRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

// Spiel (denormalisiert)

type Game struct {
	ID          string    `json:"id"`
	GameDate    string    `json:"game_date"`
	GameTime    string    `json:"game_time,omitempty"`
	HomeTeam    TeamRef   `json:"home_team"`
	AwayTeam    TeamRef   `json:"away_team"`
	Venue       VenueRef  `json:"venue"`
	League      LeagueRef `json:"league"`
	Position    string    `json:"position"`
	RefereeFee  float64   `json:"referee_fee"`
	TravelCosts float64   `json:"travel_costs"`
	KmDriven    int       `json:"km_driven"`
	Exhibition  bool      `json:"exhibition"`
	Remarks     string    `json:"remarks,omitempty"`
}
