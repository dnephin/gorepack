package repack

// Rename github.com/user/repo to example.com/newname, update the root package
// name to newname, and exclude everything in vendor
func ExampleRename() {
	err := Rename("./path/to/repo", RenameOpts{
		Imports:  map[string]string{"github.com/user/repo": "example.com/newname"},
		Package: "example.com/newname",
		Exclude:  []string{"vendor"},
	})
	if err != nil {
		panic(err)
	}
}
