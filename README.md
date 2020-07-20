Famous Firsts

Copyright 2020 by Dan Ratner

This software is covered by the MIT License, provided below for convenience.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

---

NOTES

This game is hand-written with a JavaScript front end and Golang backend. It does not use any external JavaScript or frameworks and it uses only one non-core Golang library (xid). I didn't observe any particular pattern in its design either, instead pushing for simplicity, performance, and legibility. In short, it is written in exactly the opposite way that most apps are written these days for better or for worse. I did this primarily to prove that it's still possible, though I also wanted to squeeze out every bit of performance.

For the most part the security model is quite light. That's because no sensitive data is collected, nothing is stored in a database or other persistent data store (other than log files), and the consequences of exploits are relatively modest.

Enjoy. 
Dan 
Github @dratner

---

KNOWN BUGS

* If you try to join a game with an invalid access code, you then can't make a new game (unless you Leave Game or refresh to clear things out.)

FEATURE REQUESTS

* Sarky comment if you got the right sentence exactly.
* At end of game show top sentences that fooled the most people.
* Add like button to give credit to favs even if they didn't win.
* Now that Author and Title are distinct fields, remove them from the summary field (and add them to the display function.)

WONTDO

* Give a 10 second warning. (Likely more annoying than helpful.)
* Pick how many books for a new game. (Want to keep the interface simple.)

DONE

* Randomize first vs last lines.
* Allow the gameview to be updated whenever it is blank (to avoid a single missed query)
* Show who submitted the book in the book view.
* Add commas between player names on display page
* No case sensitive access codes
* Add appropriate message for ties at the end of the game.
* Add a way to report issues through Git 
* At end of game show the list of books







