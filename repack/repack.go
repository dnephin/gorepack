/*Package repack updates package statements, and imports of those packages in
go source files
*/
package repack

import (
	"bytes"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"

	"strings"

	"github.com/pkg/errors"
)

// RenameOpts are options available to RenamePackage
type RenameOpts struct {
	// Imports is a mapping of source imports paths to target import paths
	Imports map[string]string
	// Packages is a map of relative file path to the name of the new package
	Packages map[string]string
	// Exclude is a list of relative paths to exclude from the renaming.
	Exclude []string
}

type fileRenameOpts struct {
	Canonical bool
	Package   string
	Imports   map[string]string
}

func (o RenameOpts) excludeSet() stringset {
	exclude := make(map[string]struct{}, len(o.Exclude))
	for _, path := range o.Exclude {
		exclude[path] = struct{}{}
	}
	return stringset{items: exclude}
}

func (o RenameOpts) fileOpts(dirPath string) fileRenameOpts {
	return fileRenameOpts{
		Package: o.Packages[dirPath],
		Imports: o.Imports,
	}
}

// Rename recursively reads files at path and modifies the package statement
// to a new name. Any imports of packages are also update to the new package path.
func Rename(root string, opts RenameOpts) error {
	exclude := opts.excludeSet()
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(root, path)
		switch {
		case err != nil:
			return err
		case exclude.contains(relPath) && info.IsDir():
			return filepath.SkipDir
		case exclude.contains(relPath) || info.IsDir():
			return nil
		}

		file, err := newSourcefile(path, info.Mode())
		if err != nil {
			return err
		}
		return renameInFile(file, opts.fileOpts(filepath.Dir(relPath)))
	})
}

type stringset struct {
	items map[string]struct{}
}

func (ss stringset) contains(key string) bool {
	_, ok := ss.items[key]
	return ok
}

func renameInFile(file sourcefile, opts fileRenameOpts) error {
	raw, err := scanFile(file, opts)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file.path, raw, file.mode)
	return errors.Wrapf(err, "failed to write %s", file.path)
}

type sourcefile struct {
	path string
	mode os.FileMode
	raw  []byte
}

func newSourcefile(path string, mode os.FileMode) (sourcefile, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return sourcefile{}, errors.Wrapf(err, "failed to read %s", path)
	}
	return sourcefile{
		path: path,
		mode: mode,
		raw:  raw,
	}, nil
}

func scanFile(file sourcefile, opts fileRenameOpts) ([]byte, error) {
	fset := token.NewFileSet()
	fileTokens := fset.AddFile(file.path, fset.Base(), len(file.raw))

	var err error
	errHandler := func(pos token.Position, msg string) {
		err = errors.Errorf("error scanning at %s: %s", pos, msg)
	}
	fileScanner := &scanner.Scanner{}
	fileScanner.Init(fileTokens, file.raw, errHandler, scanner.ScanComments)

	buf := newBuffer(file.raw)
	for {
		var err error
		_, tok, _ := fileScanner.Scan()
		switch {
		case tok == token.EOF:
			return buf.end(), err
		case tok == token.PACKAGE && opts.Package != "":
			err = buf.pkg(fileScanner, opts)
		case tok == token.IMPORT:
			err = buf.imports(fileScanner, opts.Imports)
		default:
			//fmt.Printf("%d: (%s) %s\n", pos, tok, literal)
		}
		if err != nil {
			return nil, err
		}
	}
}

type buffer struct {
	source []byte
	buf    *bytes.Buffer
	last   int
}

func newBuffer(source []byte) *buffer {
	return &buffer{source: source, buf: new(bytes.Buffer)}
}

func (b *buffer) end() []byte {
	b.writeToPos(len(b.source))
	return b.buf.Bytes()
}

func (b *buffer) writeToPos(pos int) {
	b.buf.Write(b.source[b.last:pos])
}

// https://golang.org/ref/spec#Package_clause
func (b *buffer) pkg(fileScanner *scanner.Scanner, opts fileRenameOpts) error {
	pos, tok, literal := fileScanner.Scan()
	if tok != token.IDENT {
		return errors.Errorf(
			"expected a package name token at %d, got (%s) %s", pos, tok, literal)
	}
	beforeTokenPos := int(pos - 1)
	b.writeToPos(beforeTokenPos)
	b.buf.WriteString(opts.Package)

	b.last = beforeTokenPos + len(literal)
	return nil
}

// https://golang.org/ref/spec#Import_declarations
func (b *buffer) imports(fileScanner *scanner.Scanner, replacements map[string]string) error {
	pos, tok, literal := fileScanner.Scan()

	switch tok {
	case token.LPAREN:
		// list of imports
		for tok != token.RPAREN && tok != token.EOF {
			pos, tok, literal = fileScanner.Scan()
			if tok == token.IDENT {
				continue
			}
			b.importStatement(pos, tok, literal, replacements)
		}
	case token.IDENT:
		// import with an alias, the next token should be the import
		return b.imports(fileScanner, replacements)
	case token.STRING:
		// single import with no alias
		b.importStatement(pos, tok, literal, replacements)
	case token.EOF:
	default:
		return errors.Errorf("expected an import token at %d, got (%s) %s", pos, tok, literal)
	}
	return nil
}

func (b *buffer) importStatement(pos token.Pos, tok token.Token, literal string, replacements map[string]string) {
	for source, target := range replacements {
		if !strings.HasPrefix(literal, `"`+source) {
			continue
		}
		beforeTokenPos := int(pos - 1)
		b.writeToPos(beforeTokenPos)
		b.buf.WriteString(strings.Replace(literal, source, target, 1))
		b.last = beforeTokenPos + len(literal)
	}

}
