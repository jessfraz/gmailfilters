package main

// filterfile defines a set of filter objects.
type filterfile struct {
	Filters []filter
}

// filter defines a filter object.
type filter struct {
	Query   string
	Archive bool
	Read    bool
	Label   string
}
