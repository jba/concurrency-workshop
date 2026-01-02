package m

// heading Interleavings

// div.flex
/* text
<div class='code'><pre>
c++
</pre></div>

is actually

<div class='code'><pre>
R0 = c
R0++
c = R0
</pre></div>
<div style="width: 30vw"></div>
*/

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
// !div.flex
