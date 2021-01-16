/*	Copyright (c) 2020, Serhat Şevki Dinçer.
	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package genius

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	gUrl       = "https://api.genius.com/"
	searchPath = "search?per_page=50&q="
	songsPath  = "artists/%d/songs?per_page=50"
	gClient    *http.Client

	gAuth   string
	tokenRE = regexp.MustCompile(`^[a-zA-Z0-9._~+/-]+=*$`) // rfc6750: oauth2 token

	spaceRE = regexp.MustCompile(`\s+`) // white space
)

// Init must be called once with a valid token (and optionally
// an existing http client) before making any genius queries.
func Init(token string, hc *http.Client) error {

	if gClient != nil {
		return nil // already initialized
	}

	if !tokenRE.MatchString(token) {
		return errors.New("genius: invalid token")
	}
	gAuth = "Bearer " + token // rfc6750: oauth2 header

	if hc != nil {
		gClient = hc
	} else {
		gClient = http.DefaultClient
	}
	return nil
}

type Artist struct {
	Id   int
	Name string
}

func (a Artist) String() string {
	return fmt.Sprintf("Artist(%d): %s", a.Id, a.Name)
}

type Song struct {
	Id    int
	Title string
}

func (s Song) String() string {
	return fmt.Sprintf("Song(%d): %s", s.Id, s.Title)
}

type gHit struct {
	Result struct {
		Primary_artist Artist
	}
}

// artist search & songs results
type gResults struct {
	Meta struct {
		Status  int
		Message string
	}
	Response struct {
		Hits  []gHit
		Songs []Song
	}
	Error_description string
}

type freq struct {
	id int
	n  uint
}

// select most occuring artist from results taking canonical name into account
func selectArtist(artist string, hits []gHit) (sel Artist) {

	// calculate frequency of artists in hit list
	var fl []freq

out:
	for i := len(hits) - 1; i >= 0; i-- {
		ar := &hits[i].Result.Primary_artist

		// ignore if canonical name is not in hit.Name
		if strings.Index(ar.Name, artist) < 0 {
			continue
		}

		for k := len(fl) - 1; k >= 0; k-- {
			if ar.Id == fl[k].id {
				fl[k].n++
				continue out
			}
		}
		fl = append(fl, freq{ar.Id, 1})
	}

	// find the highest frequency
	k := len(fl) - 1
	if k < 0 {
		return // no suitable artist found
	}
	high := fl[k]

	for k--; k >= 0; k-- {
		if high.n < fl[k].n {
			high = fl[k]
		}
	}

	// return most occurring artist
	for i := len(hits) - 1; i >= 0; i-- {
		ar := &hits[i].Result.Primary_artist

		if ar.Id == high.id {
			return *ar
		}
	}

	return // will not reach here
}

// SongsOf returns up to 50 songs of an artist
func SongsOf(artist string) (ar Artist, sl []Song, err error) {

	if gClient == nil {
		err = errors.New("genius: not initialized")
		return
	}

	// canonical artist name
	artist = strings.TrimSpace(artist)
	artist = spaceRE.ReplaceAllLiteralString(artist, " ")
	artist = strings.ToLower(artist)
	artist = strings.Title(artist)
	if len(artist) <= 0 {
		err = errors.New("genius: invalid artist")
		return
	}

	req, err := http.NewRequest("GET", gUrl+searchPath+url.QueryEscape(artist), nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", gAuth)

	// search artist
	resp, err := gClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// will store artist search & songs results
	var result gResults

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&result)
	if err != nil {
		return
	}

	if result.Meta.Status != 200 {
		msg := result.Meta.Message
		if len(msg) <= 0 {
			msg = result.Error_description
		}
		err = errors.New("genius: " + msg)
		return
	}

	// select most occuring artist from results taking canonical name into account
	ar = selectArtist(artist, result.Response.Hits)

	if ar.Id == 0 && len(ar.Name) <= 0 {
		err = errors.New("genius: no such artist found")
		return
	}

	req, err = http.NewRequest("GET", gUrl+fmt.Sprintf(songsPath, ar.Id), nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", gAuth)

	// fetch songs
	resp, err = gClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	dec = json.NewDecoder(resp.Body)
	err = dec.Decode(&result)
	if err != nil {
		return
	}

	// song list
	sl = result.Response.Songs

	if result.Meta.Status != 200 {
		err = errors.New("genius: " + result.Meta.Message)
	}
	return
}
