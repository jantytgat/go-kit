package sqr

import (
	"embed"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

// LoadQueryFromFs retrieves a query from a filesystem.
// It needs the root path to start the search from, as well as a collection name and a query name.
// The collection name equals to a direct directory name in the root path.
// The query name is the file name (without extension) to load the contents from.
// It returns and empty string and an error if the file cannot be found.
func LoadQueryFromFs(f fs.FS, rootPath, collectionName, queryName string) (string, error) {
	var err error
	var contents []byte
	switch f.(type) {
	case embed.FS:
		if contents, err = fs.ReadFile(f, path.Join(rootPath, collectionName, queryName)+".sql"); err != nil {
			return "", oopsBuilder.With("collection", collectionName).With("query", queryName).With("path", path.Join(rootPath, collectionName, queryName)+".sql").Wrap(err)
		}
	default:
		if contents, err = fs.ReadFile(f, filepath.Join(rootPath, collectionName, queryName)+".sql"); err != nil {
			return "", oopsBuilder.With("collection", collectionName).With("query", queryName).With("path", path.Join(rootPath, collectionName, queryName)+".sql").Wrap(err)
		}
	}
	return string(contents), nil
}

// loadFromFs looks for directories in the root path to create collections for.
// If a directory is found, it loads all the files in the subdirectory and adds the returned collection to the repository.
func loadFromFs(r *Repository, f fs.FS, rootPath string) error {
	if r == nil {
		return oopsBuilder.New("repository is nil")
	}

	if f == nil {
		return oopsBuilder.New("filesystem is nil")
	}

	var err error
	var files []fs.DirEntry
	if files, err = fs.ReadDir(f, rootPath); err != nil {
		return oopsBuilder.With("rootPath", rootPath).Wrap(err)
	}

	for _, file := range files {
		if file.IsDir() {
			var c collection
			if c, err = loadFilesFromDir(f, rootPath, file.Name()); err != nil {
				return err
			}

			if err = r.add(c); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadFilesFromDir loads all the files in the directory and returns a collection of queries.
func loadFilesFromDir(f fs.FS, rootPath, dirName string) (collection, error) {
	var err error
	var c = newCollection(dirName)
	var fullPath string

	switch f.(type) {
	case embed.FS:
		fullPath = path.Join(rootPath, dirName)
	default:
		fullPath = filepath.Join(rootPath, dirName)

	}

	var files []fs.DirEntry
	if files, err = fs.ReadDir(f, fullPath); err != nil {
		return c, oopsBuilder.With("path", fullPath).Wrap(err)
	}

	for _, file := range files {
		if file.IsDir() {
			return c, oopsBuilder.With("path", fullPath).With("filename", file.Name()).New("nested directories are not supported")
		}

		var contents string
		if contents, err = LoadQueryFromFs(f, rootPath, dirName, strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))); err != nil {
			return c, err
		}

		if err = c.add(strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())), contents); err != nil {
			return c, err
		}
	}
	return c, nil
}
