package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func main() {

	if err := fn.AsMain(fn.ResourceListProcessorFunc(Run)); err != nil {
		os.Exit(1)
	}
}
