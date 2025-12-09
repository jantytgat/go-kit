package sqr

// newCollection creates a new collection with the supplied name and returns it to the caller.
func newCollection(name string) collection {
	return collection{
		name:    name,
		queries: make(map[string]string),
	}
}

type collection struct {
	name    string
	queries map[string]string
}

// add adds a query to the collection.
func (c *collection) add(queryName, statement string) error {
	if _, ok := c.queries[queryName]; ok {
		return oopsBuilder.With("collection", c.name).With("query", queryName).With("statement", statement).New("query already exists")
	}
	c.queries[queryName] = statement
	return nil
}

// get retrieves a query from the collection by name.
// If the query name cannot be found, get() returns an empty string and an error.
func (c *collection) get(queryName string) (string, error) {
	if _, ok := c.queries[queryName]; !ok {
		return "", oopsBuilder.With("collection", c.name).With("query", queryName).New("query not found in collection")
	}
	return c.queries[queryName], nil
}
