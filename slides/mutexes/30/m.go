package m

////////////////////////////////////
// heading Interleavings

// cols

func f1() {
	var c int
	// code
	c++
	// !code
}

// text is actually
func f2() {
	var R0, c int
	// code
	R0 = c
	R0++
	c = R0
	// !code
}

// Make the column wider.
// html <div style="width: 25vw"></div>

// nextcol
/* text

What we want:

<div class="interleave" style="font-size: 70%">

| G1 | G2 |
| -- | -- |
| c++ |  |
|  | c++ |

</div>

What we might get:

<div class="interleave" style="font-size: 70%">

| G1 | G2 |
| -- | -- |
| R0 = c | R0 = c |
| R0++ | R0++ |
| c = R0 | c = R0 |
</div>

*/
// !cols
