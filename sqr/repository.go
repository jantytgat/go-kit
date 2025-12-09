// Package sqr enables the use of centralized storage for all SQL queries used in an application.
package sqr

import (
	"context"
	"database/sql"
	"io/fs"
	"sync"
)

// NewFromFs creates a new repository using a filesystem.
// It takes a filesystem and a root path to start loading files from and returns an error if files cannot be loaded.
func NewFromFs(f fs.FS, rootPath string) (*Repository, error) {
	repo := &Repository{
		queries: make(map[string]collection),
	}

	return repo, loadFromFs(repo, f, rootPath)
}

// A Repository stores multiple collections of queries in a map for later use.
// Queries can either be retrieved by their name, or be used to create a prepared statement.
type Repository struct {
	queries map[string]collection
	mux     sync.Mutex
}

// add adds the supplied collection to the repository.
// It returns an error if the collection already exists.
func (r *Repository) add(c collection) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	if _, ok := r.queries[c.name]; ok {
		return oopsBuilder.With("collection", c.name).New("collection already exists")
	}
	r.queries[c.name] = c
	return nil
}

// DbPrepare creates a prepared statement for the supplied database handle.
// It takes a collection name and query name to look up the query to create the prepared statement.
func (r *Repository) DbPrepare(db *sql.DB, collectionName, queryName string) (*sql.Stmt, error) {
	if db == nil {
		return nil, oopsBuilder.New("db is nil")
	}

	var err error
	var query string

	if query, err = r.Get(collectionName, queryName); err != nil {
		return nil, err
	}

	var stmt *sql.Stmt
	stmt, err = db.Prepare(query)
	return stmt, oopsBuilder.With("collection", collectionName).With("query", queryName).Wrap(err)
}

// DbPrepareContext creates a prepared statement for the supplied database handle using a context.
// It takes a collection name and query name to look up the query to create the prepared statement.
func (r *Repository) DbPrepareContext(ctx context.Context, db *sql.DB, collectionName, queryName string) (*sql.Stmt, error) {
	if db == nil {
		return nil, oopsBuilder.New("db is nil")
	}

	var err error
	var query string

	if query, err = r.Get(collectionName, queryName); err != nil {
		return nil, err
	}

	var stmt *sql.Stmt
	stmt, err = db.PrepareContext(ctx, query)
	return stmt, oopsBuilder.With("collection", collectionName).With("query", queryName).Wrap(err)
}

// Get retrieves the supplied query from the repository.
// It takes a collection name and a query name to perform the lookup and returns an empty string and an error if the query cannot be found
// in the collection.
func (r *Repository) Get(collectionName, queryName string) (string, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if s, ok := r.queries[collectionName]; ok {
		return s.get(queryName)
	}
	return "", oopsBuilder.With("collection", collectionName).With("query", queryName).New("collection not found")
}

// TxPrepare creates a prepared statement for the supplied in-progress database transaction.
// It takes a collection name and query name to look up the query to create the prepared statement.
func (r *Repository) TxPrepare(tx *sql.Tx, collectionName, queryName string) (*sql.Stmt, error) {
	if tx == nil {
		return nil, oopsBuilder.With("collection", collectionName).With("query", queryName).New("tx is nil")
	}
	var err error
	var statement string

	if statement, err = r.Get(collectionName, queryName); err != nil {
		return nil, err
	}

	var stmt *sql.Stmt
	stmt, err = tx.Prepare(statement)
	return stmt, oopsBuilder.With("collection", collectionName).With("query", queryName).Wrap(err)
}

// TxPrepareContext creates a prepared statement for the supplied in-progress database transaction using a context.
// It takes a collection name and query name to look up the query to create the prepared statement.
func (r *Repository) TxPrepareContext(ctx context.Context, tx *sql.Tx, collectionName, queryName string) (*sql.Stmt, error) {
	if tx == nil {
		return nil, oopsBuilder.New("tx is nil")
	}
	var err error
	var statement string

	if statement, err = r.Get(collectionName, queryName); err != nil {
		return nil, err
	}

	var stmt *sql.Stmt
	stmt, err = tx.PrepareContext(ctx, statement)
	return stmt, oopsBuilder.With("collection", collectionName).With("query", queryName).Wrap(err)
}
