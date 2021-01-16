/*	Copyright (c) 2020, Serhat Şevki Dinçer.
	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package main

import (
	"fmt"
	"os"

	"github.com/jfcg/genius"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`Usage: songs 'Artist Name'`)
		return
	}

	genius.Init("YOUR_TOKEN", nil)

	a, sl, e := genius.SongsOf(os.Args[1])
	if e != nil {
		fmt.Println(e)
	}

	if len(sl) > 0 {
		fmt.Println(a)
		for _, s := range sl {
			fmt.Println(s)
		}
	}
}
