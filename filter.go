package main

// filterfile defines a set of Filter objects.
type filterfile struct {
	Filters []Filter
}

// Filter defines a filter object.
type Filter struct {
	Query   string
	Archive bool
	Read    bool
	Label   string
}
