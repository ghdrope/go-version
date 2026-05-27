package main

import (
	"fmt"

	"github.com/ghdrope/go-version"
)

func main() {
	fmt.Println("=== String() ===")
	fmt.Println(version.String())
	fmt.Println()

	fmt.Println("=== Short() ===")
	fmt.Println(version.Short())
}
