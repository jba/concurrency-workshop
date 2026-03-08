package testdata

// heading Elide Test

// code
func example() {
	x := 1
	// elide
	y := 2
	z := 3
	// !elide
	fmt.Println(x)
}
// !code
