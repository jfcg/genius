/*	Copyright (c) 2020, Serhat Şevki Dinçer.
	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package genius

import (
	"testing"
)

func TestInit(t *testing.T) {
	// bad tokens
	bad := [...]string{"", "a*", "bş"}

	for i := len(bad) - 1; i >= 0; i-- {
		e := Init(bad[i], nil)

		if e == nil || gClient != nil {
			t.Fatal("accepted bad token")
		}
	}
}

func TestSongs(t *testing.T) {
	a, sl, e := SongsOf("jackson")

	if e == nil || e.Error() != "genius: not initialized" ||
		len(sl) > 0 || len(a.Name) > 0 {
		t.Fatal("wrong error before initialization")
	}

	Init("INVALID_TOKEN", nil)

	// bad artists
	bad := [...]string{"", " \t"}

	for i := len(bad) - 1; i >= 0; i-- {
		a, sl, e = SongsOf(bad[i])

		if e == nil || e.Error() != "genius: invalid artist" ||
			len(sl) > 0 || len(a.Name) > 0 {
			t.Fatal("accepted bad artist name")
		}
	}

	a, sl, e = SongsOf("jackson")

	if e == nil || len(sl) > 0 || len(a.Name) > 0 {
		t.Fatal("query with invalid token")
	}
}
