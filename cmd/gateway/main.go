package main

import (
	"fmt"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
)

func main() {
	d := dmp.New()
	diffs := d.DiffMain("Hello World", "HelloWorld!!!", false)
	fmt.Println(diffs)
	patches := d.PatchMake(diffs)
	result, _ := d.PatchApply(patches, "Hello World")
	fmt.Println(result) // "Hello Go World"
}
