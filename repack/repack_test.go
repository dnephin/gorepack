package repack

import (
	"io/ioutil"
	"testing"

	"go/scanner"
	"go/token"

	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/golden"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRename(t *testing.T) {
	tmp := fs.NewDir(t, "test-rename", fs.FromDir(golden.Path("pkgsource")))
	defer tmp.Remove()
	opts := RenameOpts{
		Imports: map[string]string{
			"example.com/user/pkgsource": "vanity.fake/newsy",
		},
		Package: "vanity.fake/newsy",
	}

	err := Rename(tmp.Path(), opts)
	require.NoError(t, err)

	golden.Assert(t, content(t, tmp.Join("file.go")), "test-rename-expected/file.go")
	golden.Assert(t, content(t, tmp.Join("cmd/foo/main.go")), "test-rename-expected/cmd/foo/main.go")
	golden.Assert(t, content(t, tmp.Join("util/util.go")), "test-rename-expected/util/util.go")
	golden.Assert(t,
		content(t, tmp.Join("util/sub/subutil.go")),
		"test-rename-expected/util/sub/subutil.go")
}

func content(t require.TestingT, path string) string {
	raw, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	return string(raw)
}

func TestBufferImports(t *testing.T) {
	var testcases = []struct {
		doc      string
		source   string
		expected string
	}{
		{
			doc:      "single import no alias with replacement",
			source:   `"replace.this/import"`,
			expected: `"example.com/new/path"`,
		},
		{
			doc:      "single import prefix replacement",
			source:   `"replace.this/import/sub/foo"`,
			expected: `"example.com/new/path/sub/foo"`,
		},
		{
			doc:      "single import no alias no replacement",
			source:   `"example.com/something"`,
			expected: `"example.com/something"`,
		},
		{
			doc:      "single import with alias and replacement",
			source:   `myalias "replace.this/import"`,
			expected: `myalias "example.com/new/path"`,
		},
		{
			doc: "list of imports",
			source: `(
	"replace.this/import"
	foo "replace.this/other"
	nope "not/this/one"
	"not/this/one/either"
)`,
			expected: `(
	"example.com/new/path"
	foo "example.com/new/other"
	nope "not/this/one"
	"not/this/one/either"
)`,
		},
	}
	replacements := map[string]string{
		"replace.this/import": "example.com/new/path",
		"replace.this/other":  "example.com/new/other",
	}

	for _, testcase := range testcases {
		t.Run(testcase.doc, func(t *testing.T) {
			buf := newBuffer([]byte(testcase.source))
			scanr := newBytesScanner(t, buf.source)

			err := buf.imports(scanr, replacements)
			require.NoError(t, err)
			assert.Equal(t, testcase.expected, string(buf.end()))
		})
	}
}

func newBytesScanner(t *testing.T, source []byte) *scanner.Scanner {
	fset := token.NewFileSet()
	fileTokens := fset.AddFile("", fset.Base(), len(source))

	errHandler := func(pos token.Position, msg string) {
		t.Fatalf("error scanning at %s: %s", pos, msg)
	}
	fileScanner := &scanner.Scanner{}
	fileScanner.Init(fileTokens, source, errHandler, scanner.ScanComments)
	return fileScanner
}
