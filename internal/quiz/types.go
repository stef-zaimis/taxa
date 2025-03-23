package quiz

// A taxonomic unit used for quiz options
type Taxon struct {
	TaxonID string `json:"taxonId,omitempty"`
	ScientificName string `json:"scientificName"`
	Authorship string `json:"authorship"`
	GBIFKey string `json:"gbifKey,omitempty"`
	Rank string `json:"rank"`
	HasMedia bool `json:"hasMedia"`

	// Optional but adding it for now
	Status string `json:"status"`

	Kingdom string `json:"kingdom,omitempty"`
	Phylum string `json:"phylum,omitempty"`
	Class string `json:"class,omitempty"`
	Order string `json:"order,omitempty"`
	SuperFamily string `json:"superFamily,omitempty"`
	Family string `json:"family,omitempty"`
	SubFamily string `json:"subFamily,omitempty"`
	Tribe string `json:"tribe,omitempty"`
}

// A single quiz question instance
type Question struct {
	ImageURL string `json:"imageUrl"`
	Options []Taxon `json:"options"`
	CorrectIndex int `json:"correctIndex"`
	CorrectAnswer Taxon `json:"correctAnswer"`
}
