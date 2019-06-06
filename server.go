package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// serveEtudes serves etude midi files from the current working directory.
func serveEtudes(hostport string, maxAgeSeconds int, midijsPath string) {
	var err error
	err = mkWebPages()
	if err != nil {
		log.Fatalf("could not write web pages: %v", err)
	}
	err = validDirPath(midijsPath)
	if err != nil {
		log.Fatalf("invalid midijs path: %v", err)
	}
	os.Setenv("MIDIJS", midijsPath)
	defer os.Unsetenv("MIDIJS")
	os.Setenv("ETUDE_MAX_AGE", fmt.Sprintf("%d", maxAgeSeconds))
	defer os.Unsetenv("ETUDE_MAX_AGE")
	http.Handle("/", http.HandlerFunc(indexHndlr))
	http.Handle("/etude/", http.HandlerFunc(etudeHndlr))
	http.Handle("/midijs/", http.HandlerFunc(midijsHndlr))
	log.Printf("midijs path is %s", os.Getenv("MIDIJS"))
	var serveSecure bool
	var certpath, certkeypath string
	if hostport == ":443" {
		certpath, certkeypath, err = getCertPaths()
		if err != nil {
			log.Printf("Can't find SSL certificates: %v", err)
			hostport = ":80"
		}
		serveSecure = true
	}
	log.Printf("serving on %s\n", hostport)
	switch serveSecure {
	case true:
		if err := http.ListenAndServeTLS(hostport, certpath, certkeypath, nil); err != nil {
			log.Fatalf("Could not listen on port %s : %v", hostport, err)
		}
	default:
		if err := http.ListenAndServe(hostport, nil); err != nil {
			log.Fatalf("Could not listen on port %s : %v", hostport, err)
		}
	}
}

// getCert attempts to retrieve a certficate and key for use with
// ListenAndServeTLS. It returns an error if either item cannot be found but
// does not otherwise attempt to validate them. That is left up to
// ListenAndServeTLS.
func getCertPaths() (certpath string, keypath string, err error) {
	certpath = os.Getenv("IETUDE_CERT_PATH")
	if certpath == "" {
		err = fmt.Errorf("no environment variable IETUDE_CERT_PATH")
		return
	}
	keypath = os.Getenv("IETUDE_CERTKEY_PATH")
	if keypath == "" {
		err = fmt.Errorf("no environment variable IETUDE_CERTKEY_PATH")
		return
	}
	return
}

// indexHndlr returns index.html
func indexHndlr(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// midijsHndlr returns files from the MIDIJS directory.
func midijsHndlr(w http.ResponseWriter, r *http.Request) {
	what := strings.Split(r.URL.Path, "/")
	if what[1] != "midijs" {
		log.Fatalf("programming error. got request path that didn't start with 'midijs': %s", r.URL.Path)
	}

	dir := os.Getenv("MIDIJS")
	pathelements := append([]string{dir}, what[2:]...)
	path := filepath.Join(pathelements...)
	http.ServeFile(w, r, path)

}

// etudeHndlr returns a midi file that matches the get request or a 404 for
// incorrectly specified etudes. The pattern is
// /etude/<key>/<scale>/<instrument> where key is a pitchname like "c" or
// "aflat", <scale> is a scalename like "pentatonic", and instrument is a
// formatted General Midi instrument name like "acoustic_grand_piano".  If any
// of the forgoing are unknown or unsupported by this app, etudeHndlr gives a
// 400 response (StatusBadRequest). If the request matches a valid filename,
// the file will be returned in the response body if it exists and is younger
// than the maximum age imposed by this service. Otherwise the app will
// generate it so it can be returned.
func etudeHndlr(w http.ResponseWriter, r *http.Request) {
	what := strings.Split(r.URL.Path, "/")
	if what[1] != "etude" {
		log.Fatalf("programming error. got request path that didn't start with 'etude': %s", r.URL.Path)
	}
	if !validEtudeRequest(what[2:]) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	filename := strings.Join(what[2:], "_") + ".mid"
	makeEtudesIfNeeded(filename, what[4])
	http.ServeFile(w, r, filename)
	// log the request in format that's convenient for analysis
	log.Printf("%s %s %s %s %s\n", r.RemoteAddr, what[2], what[3], what[4], "served")
}

// makeEtudesIfNeeded generates a full set of etudes in the current
// working directory if the requested file doesn't exist or is older
// than the age limit set by serveEtudes in os.Environ. Otherwise
// it does nothing.
func makeEtudesIfNeeded(filename, instrumentName string) {
	var exists = true // initial assumption
	finfo, err := os.Stat(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			// something's really wrong
			log.Fatalf("error statting %s: %v", filename, err)
		}
		exists = false
	}
	// if file exists, we need to check its age
	if exists {
		maxage, e := strconv.Atoi(os.Getenv("ETUDE_MAX_AGE"))
		if e != nil {
			log.Fatalf("programming error. can't convert ETUDE_MAX_AGE '%s' to integer", os.Getenv("ETUDE_MAX_AGE"))
		}

		maxduration := time.Duration(maxage) * time.Second
		modtime := finfo.ModTime()
		if time.Now().Sub(modtime) < maxduration {
			// nothing to do
			return
		}
	}
	// need to generate if we get to here
	iInfo, _ := getSupportedInstrumentByName(instrumentName) // already validated. ignore err value
	// fmt.Printf("%v %s\n", iInfo, filename)
	instrument := iInfo.gmnumber - 1
	midilo := iInfo.midilo
	midihi := iInfo.midihi
	tempo := 120
	mkAllEtudes(midilo, midihi, tempo, instrument, iInfo.name)
}

// validEtudeRequest returns true if the request is correctly formed
// and references a valid etude file.
func validEtudeRequest(ksi []string) (ok bool) {
	if len(ksi) != 3 {
		return
	}
	if !validKeyName(ksi[0]) {
		return
	}
	if !validScaleName(ksi[1]) {
		return
	}
	if !validInstrumentName(ksi[2]) {
		return
	}

	ok = true
	return
}

type nameInfo struct {
	fileName string
	uiName   string
}

var keyInfo = []nameInfo{
	{"c", "C"},
	{"dflat", "D♭"},
	{"d", "D"},
	{"eflat", "E♭"},
	{"e", "E"},
	{"f", "F"},
	{"gflat", "G♭"},
	{"g", "G"},
	{"aflat", "A♭"},
	{"a", "A"},
	{"bflat", "B♭"},
	{"b", "B"},
}

// validKeyName returns true if the key name is in the ones we support.
func validKeyName(name string) (ok bool) {
	for _, k := range keyInfo {
		if k.fileName == name {
			ok = true
			break
		}
	}
	return
}

var scaleInfo = []nameInfo{
	{"pentatonic", "Pentatonic"},
	{"final", "Chromatic Final"},
	{"plus_four", "Plus Four"},
	{"plus_seven", "Plus Seven"},
	{"four_and_seven", "Four and Seven"},
	{"raised_five", "Harmonic Minor 1"},
	{"raised_five_with_four_or_seven", "Harmonic Minor 2"},
}

// validScaleName returns true if the scale name is in the ones we support.
func validScaleName(name string) (ok bool) {
	for _, s := range scaleInfo {
		if s.fileName == name {
			ok = true
			break
		}
	}
	return
}

// validInstrumentName returns true if the instrument name is in the ones we
// support.
func validInstrumentName(name string) (ok bool) {
	_, err := getSupportedInstrumentByName(name)
	if err == nil {
		ok = true
	}
	return
}
