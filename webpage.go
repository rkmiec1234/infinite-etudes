package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	. "github.com/Michael-F-Ellis/goht" // dot import makes sense here
)

const (
	crossMark string = "&#x2717;"
	checkMark string = "&#x2713;"
)

type SilenceOption struct {
	value int    // binary mask. 1-bits are silent
	html  string // three circles (white or green) indicate which repeats are silent.
}

var silencePatterns = []SilenceOption{
	{0, checkMark + checkMark + checkMark},
	{1, checkMark + checkMark + crossMark},
	{2, checkMark + crossMark + checkMark},
	{4, crossMark + checkMark + checkMark},
	{3, checkMark + crossMark + crossMark},
	{5, crossMark + checkMark + crossMark},
	{6, crossMark + crossMark + checkMark},
	{7, crossMark + crossMark + crossMark},
}

// mkWebPages constructs the application web pages in the current working
// directory.
func mkWebPages() (err error) {
	err = mkIndex()
	return
}

func mkIndex() (err error) {
	var buf bytes.Buffer
	// <head>
	head := Head("",
		Meta(`name="viewport" content="width=device-width, initial-scale=1"`),
		Meta(`name="description", content="Infinite Etudes demo"`),
		Meta(`name="keywords", content="music,notation,midi,tbon"`),
		Link(`rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css"`),
		indexCSS(),
		indexJS(), // js for this page
		// js midi libraries
		Script("src=/midijs/libtimidity.js charset=UTF-8"),
		Script("src=/midijs/midi.js charset=UTF-8"),
	)

	// <html>
	page := Html("", head, indexBody())
	err = Render(page, &buf, 0)
	if err != nil {
		return
	}
	err = ioutil.WriteFile("index.html", buf.Bytes(), 0644)
	return
}

func indexBody() (body *HtmlTree) {
	header := Div(`style="text-align:center; margin-bottom:2vh;"`,
		H2("class=title", "Infinite Etudes"),
		Em("", "Ear training for your fingers"),
	)
	// Etude menus:
	// Scale pattern
	var scales []interface{}
	for _, ptn := range patternInfo { // scaleInfo is defined in server.go
		value := fmt.Sprintf(`value="%s"`, ptn.fileName)
		scales = append(scales, Option(value, ptn.uiName))
	}
	scaleSelect := Div(`class="Column" id=scale-div`, Label(``, "Pattern", Select("id=scale-select", scales...)))
	// scaleSelectLabel := Label(`class=sel-label`, "Pattern")

	// Key
	var keys []interface{}
	for _, k := range keyInfo {
		value := fmt.Sprintf(`value="%s" aria-label="%s"`, k.fileName, k.uiAria)
		keys = append(keys, Option(value, k.uiName))
	}
	keySelect := Div(`class="Column" id="key-div"`, Label(``, "Tonic", Select("id=key-select", keys...)))
	// Interval1 and Interval2
	var intervals []interface{}
	for _, v := range intervalInfo {
		value := fmt.Sprintf(`value="%s" aria-label="%s"`, v.fileName, v.uiAria)
		uival := fmt.Sprintf("%d (%s)", v.size, v.uiName)
		intervals = append(intervals, Option(value, uival))
	}
	interval1Select := Div(`class="Column" id="interval1-div"`, Label(``, "Interval 1", Select("id=interval1-select", intervals...)))
	interval2Select := Div(`class="Column" id="interval2-div"`, Label(``, "Interval 2", Select("id=interval2-select", intervals...)))
	interval3Select := Div(`class="Column" id="interval3-div"`, Label(``, "Interval 3", Select("id=interval3-select", intervals...)))
	// Instrument sound
	var sounds []interface{}
	for _, iinfo := range supportedInstruments {
		name := iinfo.displayName
		value := fmt.Sprintf(`value="%s"`, iinfo.name)
		sounds = append(sounds, Option(value, name))
	}
	soundSelect := Div(`class="Column" id="sound-div"`, Label(``, "Instrument", Select("id=sound-select", sounds...)))

	// Metronome
	var metros []interface{}
	for _, ptn := range []string{"on", "downbeat", "off"} {
		attrs := fmt.Sprintf(`value="%s"`, ptn)
		metros = append(metros, Option(attrs, ptn))
	}
	metroSelect := Div(`class="Column" id="metro-div"`, Label(``, "Metronome", Select("id=metro-select", metros...))) // Metronome control

	var tempos []interface{}
	var tempoValues []int
	for i := 60; i < 484; i += 4 {
		tempoValues = append(tempoValues, i)
	}
	for _, bpm := range tempoValues {
		name := fmt.Sprintf("%d", bpm)
		value := fmt.Sprintf(`value="%d"`, bpm)
		if bpm == 120 {
			value += " selected" // use 120 as the default value
		}
		tempos = append(tempos, Option(value, name))
	}
	tempoSelect := Div(`class="Column" id="tempo-div"`, Label(``, "Tempo", Select("id=tempo-select", tempos...)))

	// Repeats
	var repeats []interface{}
	for _, reps := range []string{"3", "2", "1", "0"} {
		attrs := fmt.Sprintf(`value="%s"`, reps)
		repeats = append(repeats, Option(attrs, reps))
	}
	repeatSelect := Div(`class="Column" id="repeat-div"`, Label(``, "Repeats", Select("id=repeat-select", repeats...)))

	// Silences
	var silences []interface{}
	for _, ptn := range silencePatterns {
		attrs := fmt.Sprintf(`value="%d"`, ptn.value)
		silences = append(silences, Option(attrs, ptn.html))
	}
	silenceSelect := Div(`class="Column" id="silence-div"`, Label(``, "Muting", Select("id=silence-select", silences...)))

	// Controls
	playBtn := Button(`onclick="playStart()"`, "Play")
	stopBtn := Button(`onclick="playStop()"`, "Stop")
	downloadBtn := Button(`onclick="downloadEtude()"`, "Download")

	// Assemble everything into the body element.
	body = Body("", header,
		Div(`class="Row" id="scale-row"`, scaleSelect, keySelect, interval1Select, interval2Select, interval3Select),
		Div(`class="Row"`, soundSelect, metroSelect),
		Div(`class="Row"`, tempoSelect, repeatSelect, silenceSelect),
		Div(`style="padding-top:1vh;"`, playBtn, stopBtn, downloadBtn),
		quickStart(),
		forTheCurious(),
		forVocalists(),
		controls(),
		custom(),
		variations(),
		// faq(),
		biblio(),
		coda(),
	)
	return
}

func quickStart() (div *HtmlTree) {
	div = Div("",
		H3("", "For the impatient"),
		Ol("",
			Li("", `Choose a pattern,`),
			Li("", `Choose a tonic note (or a set of intervals),`),
			Li("", `Choose an instrument sound,`),
			Li("", `Click 'Play' and play along.`),
		),
	)
	return
}

func forTheCurious() (div *HtmlTree) {
	heading := "For the curious"
	p0 := `Infinite Etudes generates ear/finger training etudes for instrumentalists.
	The emphasis is on improving your ability to play what you hear by thoroughly exploring
	all the combinations of 2, 3 and 4 pitches over the full range of your instrument.`

	p1 := `The etudes follow a simple four
	bar form: a sequence of different notes is played on beats 1, 2, and 3
	and a rest on beat 4. By default, each bar is played four times before moving on -- so
	you have 3 chances to play the sequence after the first hearing. You can control
	the number of repeats.  You can also choose to silence one or more of the repeated
	measure.
	`
	p2 := `The program is called 'Infinite Etudes' because the number of
	possible orderings of the sequences easily exceeds the number of stars in
	the universe, i.e. you'll never play the same etude twice. Luckily, the
	goal is to learn to recognize and play the individual sequences. That
	turns out to be a much more reasonable task (and the infinite sequence
	orderings are actually helpful because they prevent you from relying on
	muscle memory.)
    `

	p14 := `<strong>One Interval</strong> presents 12 instances of the same
	interval pair, i.e. 3 notes, in random order. Each instance begins on a
	different pitch so that all 12 pitches are covered. <em>Note: For brevity, the
	score examples shown here are captured with a repeat count of zero.</em>`

	p15 := `<strong>Tonic Intervals</strong> presents 13 different
	intervals, i.e., all possible pitches relative to the chosen tonic
	pitch. Use pattern is as a self-test to gauge your progress.`

	p16 := `<strong>Two Intervals</strong> is, as you might expect, a series
	of three pitches specified by the interval1 and interval2 selectors. The
	score example shows a typical etude produced by choosing 4 half steps (a
	major third) for the lower interval and 3 half steps (a minor third) for
	the upper interval, i.e. a a major triad in root position. It's important
	to be able to recognize and play the notes of any interval pattern in any
	order. For 3 notes, there are 6 possible orderings and the program
	arranges for each ordering to occur twice among the 12 sequences
	presented.`

	p17 := `<strong>Three Intervals</strong> is similar to the Two Interval
	pattern but uses 3 intervals to produce 4-note sequence. There are 24
	examples in etudes produced with this pattern because you can play 4
	notes in 24 different orders. The example shows a typical etude
	constructed with a 2-2-1 pattern of half steps, corresponding to the
	first 4 notes of a major scale.`

	div = Div("",
		H3("", heading),
		P("", p0),
		P("", p1),
		P("", p2),
		H3("", "Etude Patterns"),
		H4("", "One Interval"),
		P("", p14),
		Img(`src="img/one_interval_excerpt.png" class="example"`),
		H4("", "Tonic Intervals"),
		P("", p15),
		Img(`src="img/allintervals_excerpt.png" class="example"`),
		H4("", "Two Intervals"),
		P("", p16),
		Img(`src="img/two_interval_excerpt.png" class="example"`),
		H4("", "Three Intervals"),
		P("", p17),
		Img(`src="img/three_interval_excerpt.png" class="example"`),
	)
	return
}
func forVocalists() (div *HtmlTree) {
	p1 := `I conceived Infinite Etudes as an aid for instrumentalists. I've since
	found it's also quite useful as a daily vocal workout for intonation. The
	instrument selection menu has choir ahh sounds for soprano, alto, tenor and bass ranges.`
	div = Div("",
		H3("", "For Vocalists"),
		P("", p1),
	)
	return
}

func controls() (div *HtmlTree) {
	p1 := `The Pattern selector allows you to choose one of the patterns
	described above. Your choice affects the visibility of the interval and
	tonal center selectors. The interval selectors, (Interval1, Interval2 and
	Interval3), appear according to the number intervals in the chosen
	pattern. The interval choices are labeled by the number of half steps and
	the corresponding musical name, e.g. "4 (Minor Third)". The Tonal Center
	selector appears only when the Tonic Intervals pattern is selected.`

	p2 := `The Instrument selector provides a choice of common instrument sounds. Your choice also
	determines the range of pitches that can occur within an etude.`

	p2a := `Each etude starts on a randomly selected pitch somewhere between
	the lowest and highest notes that can commonly be played on your chosen
	instrument. The sequence orderings are a random walk constructed to that
	the first pitch of each sequence is "close" to the preceding pitch
	without wandering outside the playable range of your instrument.`

	p3 := `By default the metronome gives an initial 1 measure count-in and
	continues to click on each beat of the etude.  You can control this with the
	Metronome selector. Choose "downbeat" to have it click only on beat 1 of each measure.
	Choose "off" for silence after the count-in.`

	p4 := `Infinite Etudes generates MIDI files in 4/4 time with the tempo
	defaulted to 120 beats per minute. If you need it slower or faster, use
	the Tempo selector to choose a value between 60 and 480 beats per
	minute.`

	p5 := `Use the Repeats selector to change the number of repeats for each sequence. The default is
	3. You can set it to 2 or 1 to increase the challenge. You can also set it to 0, but that's
	not useful unless you want to download an example to import into a score editor.`

	p6 := `The Muting selector allows you silence one or more of the repeated
	measures. The cross mark symbol, &#x2717;, indicates a silent measure and
	the check mark, &#x2713;, indicates an audible one.`

	div = Div("",
		H3("", "Controls"),
		H4("", "Pattern"),
		P("", p1),
		H4("", "Instrument"),
		P("", p2),
		P("", p2a),
		H4("", "Metronome"),
		P("", p3),
		H4("", "Tempo"),
		P("", p4),
		H4("", "Repeats"),
		P("", p5),
		H4("", "Muting"),
		P("", p6),
	)
	return
}
func custom() (div *HtmlTree) {
	p1 := `If you need something beyond the available tempi and instrument sounds, the easiest solution
	is to use the download button to save a local copy of a file and play it with
	a program that allows you finer control of the playback.  I recommend QMidi for Mac. I don't
	know what's good on PC but a little Googling should turn up something appropriate. Downloading
	also allows you to play the files through better equipment for more realistic sound.`

	p2 := `You might also consider installing MuseScore, the excellent open
	source notation editor. Version 3.1 and higher does a very good job
	importing Infinite Etudes midi files. Besides controlling tempo, you can
	print the etude as sheet music or play it back with real-time
	highlighting of each note as it's played.`

	p3 := `A third option, if you have software skills, is to install
	Infinite Etudes on your computer from the source code on <a
	href="https://github.com/Michael-F-Ellis/infinite-etudes">GitHub.</a> and
	adapt the program to your needs.`

	div = Div("",
		H3("", "Customizing"),
		P("", p1),
		P("", p2),
		P("", p3),
	)
	return
}
func variations() (div *HtmlTree) {
	p1 := `As you progress, some sequences will become easy to recognize and
	play before others. When you nail a particular sequence correctly and
	confidently on first hearing (hooray!), you can put the remaining two bars
	to good use in a variety of ways. Here are a few suggestions, some simple
	and some difficult:`
	var variants = []string{
		`Finger it differently.`,
		`Change the bowing or picking.`,
		`Play it with the other hand (keyboards).`,
		`Play it in the same octave on different strings (string instruments).`,
		`Play one note up or down an octave.`,
		`Play the whole sequence up or down an octave.`,
		`Play it in both hands one or more octaves apart.`,
		`Play it up or down a fifth (or fourth, third, ...).`,
		`Play it as a chord.`,
		`Find a bass note or chord that works with the sequence.`,
		`Mess with the rhythm, accents, dynamics, timbre, ...`,
		`Shred it in sixteenth note cross-rhythm, e.g. 1231 2312 3123`,
		`Fill in between the notes.`,
		`Invent a counter-melody,`,
		`or simply take a deep breath and relax your fingers.`,
	}
	p2 := `Above all, make some music whenever possible!`
	// need []interface{} to pass strings as Li elements to Ul()
	var ivariants []interface{}
	for _, s := range variants {
		ivariants = append(ivariants, Li("", s))
	}
	div = Div("",
		H3("", "Variations"),
		P("", p1),
		Ul("", ivariants...),
		P("", p2),
	)
	return
}

/*
func faq() (div *HtmlTree) {
	qa := func(q string, a ...string) (div *HtmlTree) {
		var item []interface{}
		item = append(item, P("", Strong("", Em("", q))))
		for _, s := range a {
			item = append(item, P("", s))
		}
		div = Div("", item...)
		return
	}
	q1 := qa(`Why 3 notes rather than 4 or 5 ...?`,
		`The math is less friendly for longer sequences. There are 11880 possible sequences of 4 notes and
	95040 sequences of 5 notes. You can get through all 1320 3-note sequences every 12 days in 15 minutes/day playing the
	Chromatic Final etudes. To do that with 4-notes sequences would take over 3 months (108 days).`,
		`With 5-note sequences it would take more than 2 years.`)

	q2 := qa(`Are 3 notes enough to be of benefit?`,
		`My own experience says 'yes'. At the piano, I've experienced
	very noticeable improvement in my ability to play by ear as well as in my sight-reading. I attribute both
	to having to devote less mental effort to fingering.`,
		`As a singer, I use the etudes as a daily exercise to work on intonation through my full vocal range.`)

	div = Div("",
		H3("", `FAQ`),
		q1,
		q2,
	)
	return
}
*/
func coda() (div *HtmlTree) {
	p1 := `I wrote Infinite Etudes for two reasons: First, as a tool for my
	own practice on piano and viola; second as a small project to develop a
	complete application in the Go programming language. I'm happy with it on
	both counts and I hope you find it useful also. The source code is available
	on <a href="https://github.com/Michael-F-Ellis/infinite-etudes">GitHub.</a>
	<br><br>Mike Ellis<br>Weaverville NC<br>May 2019`
	div = Div("",
		H3("", "Coda"),
		P("", p1),
	)
	return
}

func cite(citation, comment string) (div *HtmlTree) {
	div = Div("",
		P("", "<em>"+citation+"</em>"),
		P("", "<small>"+comment+"</small>"),
	)
	return
}

func biblio() (div *HtmlTree) {
	div = Div("",
		H3("", "References"),
		P("", "A few good books that influenced the development of Infinite Etudes:"),

		cite(`Brown, Peter C. Make It Stick : the Science of Successful
        Learning. Cambridge, Massachusetts :The Belknap Press of Harvard University
        Press, 2014.`,
			`An exceedingly readable and practical summary of what works and doesn't
	   work for efficient learning.  I've attempted to incorporate the core
	   principles (frequent low stakes testing, interleaving and spaced
	   repetition) into the design of Infinite Etudes.`),

		cite(`Huron, David. Sweet Anticipation: Music and the Psychology of
		Expectation.  Cambridge, Massachusetts : The MIT Press, 2006`,
			`The book's central theme is a theory that a large part of what makes
		music enjoyable is a combination of satisfaction from predicting what
		comes next and delight when our predictions are occasionally confounded
		in interesting ways. Regarding the development of Infinite Etudes,
		Chapter 4 "Auditory Learning" and Chapter 10 "Tonality" were quite
		useful.`),

		cite(`Werner, Kenny. Effortless Mastery : Liberating the Master Musician Within. Innermusic Publishing, 2011`,
			`Jazz pianist Kenny Werner's autobiographical take on his own road to
		mastery through mindfulness.  The title refers the sensation of
		effortlessness that accompanies mastery rather than to some magical
		method of learning without practicing. I took from it a sense of the
		value of patience in allowing musicianship to develop.`),

		cite(`Wooten, Victor L. The Music Lesson : a Spiritual Search for
		Growth Through Music. New York, New York : The Berkley Publishing
		Group, 2008.`,
			`Pearls of musical wisdom are threaded throughout Grammy Award winning
		bassist Victor Wooten's fanciful tale of adventures with music teachers
		who show up unannounced at his home (think Carlos Castaneda without the
		hallucinogens.) Read it as a counterpoint to the emphasis on notes
		embodied in Infinite Etudes. Good music making is also about
		articulation, technique, emotion, dynamics, tempo, tone, phrasing and
		space, or as one of Wooten's teachers puts it: "Never lose the groove
		in order to find a note!"`),
	)
	return
}

func indexCSS() *HtmlTree {
	return Style("", `
    body {
	  margin: 0;
	  height: 100%;
	  overflow: auto;
	  background-color: #000;
	  color: #CFC;
	  }
    h1 {font-size: 300%; margin-bottom: 1vh}
    h2 {font-size: 200%}
	h2.title {text-align: center;}
    h3 {font-size: 150%; margin-left: 2vw}
    h4 {
        font-size: 120%;
        margin-left: 2vw;
        margin-top: 1vw;
        margin-bottom: 1vw;
    }
    p {
        font-size: 100%;
        margin-left: 5%;
        margin-right: 10%;
        margin-top: 1%;
        margin-bottom: 5%;
    }
    img.example {
        margin-left: 5%;
        margin-right: 10%;
        width: 85vw;
	}
	label {
		display: inline-block;
		text-align: center;
		font-size: 80%;
	}
    select {
	  display: inline-block;
	  font-size: 125%;
	  margin-bottom: 1%;
	  background-color: white;
	}
    button {
	  margin-left: 5%;
	  margin-bottom: 1%;
	  background-color: #ADA;
	}
    a {font-size: 100%}
    button.nav {
        font-size: 120%;
        margin-right: 1%;
        background-color: #DFD;
    }
    input {font-size: 100%}
    li {
        font-size: 100%;
        margin-left: 5%;
        margin-right: 10%;
        margin-bottom: 0.5%;
    }
    pre {font-size: 75%; margin-left: 5%}
	/* hover color for buttons */
    input[type=submit]:hover {background-color: #0a0}
	input[type=button]:hover {background-color: #0a0}
	/* */
	.Row {
    display: table;
    width: auto;
    table-layout: auto;
    border-spacing: 10px;
    }
    .Column{
    display: table-cell;
    /* background-color: red; */
    }
	`)
}

func indexJS() (script *HtmlTree) {
	script = Script("",
		`
		// chores at start-up
		function start() {
		  // Chrome and other browsers now disallow AudioContext until
		  // after a user action.
		  document.body.addEventListener("click", MIDIjs.resumeAudioContext);
		  var scaleselect = document.getElementById("scale-select")
		  scaleselect.addEventListener("change", manageInputs)
		  manageInputs()
		}
		// returns true if the selected key is an interval name
		function isIntervalName(name) {
			var inames = ['minor2', 'major2', 'minor3', 'major3', 'perfect4', 'tritone',
			'perfect5', 'minor6', 'major6', 'minor7', 'major7', 'octave']
			return inames.includes(name)
		}
		// manageInputs adjusts the enable status of the key and interval widgets
		// when scale-select value changes
		function manageInputs() {
			var key = document.getElementById("key-div")
			var interval1 = document.getElementById("interval1-div")
			var interval2 = document.getElementById("interval2-div")
			var interval3 = document.getElementById("interval3-div")
			var scalePattern = document.getElementById("scale-select").value
			if (scalePattern == "interval") {
				interval1.style.display=""
				interval2.style.display="none"
				interval3.style.display="none"
				key.style.display="none"
				return
			}
			if (scalePattern == "intervalpair") {
				interval1.style.display=""
				interval2.style.display=""
				interval3.style.display="none"
				key.style.display="none"
				return
			}
			if (scalePattern == "intervaltriple") {
				interval1.style.display=""
				interval2.style.display=""
				interval3.style.display=""
				key.style.display="none"
				return
			}
			// all the other patterns are chosen by key
			interval1.style.display="none"
			interval2.style.display="none"
			interval3.style.display="none"
			key.style.display=""
			return
		}
		// Read the selects and return the URL for the etude to be played or downloaded.
		function etudeURL() {
		  scale = document.getElementById("scale-select").value
		  key = document.getElementById("key-select").value
		  if (scale != "intervals" &&  isIntervalName(key)) {
			  alert(key + " is only valid when the scale pattern is Intervals.")
			  return ""
		  }
		  if (key=="random") {
			  key=randomKey()
			  };
		  interval1 = document.getElementById("interval1-select").value
		  interval2 = document.getElementById("interval2-select").value
		  interval3 = document.getElementById("interval3-select").value
		  sound = document.getElementById("sound-select").value
		  metronome = document.getElementById("metro-select").value
		  tempo = document.getElementById("tempo-select").value
		  repeats = document.getElementById("repeat-select").value
		  silent = document.getElementById("silence-select").value
		  return "/etude/" + key + "/" + scale + "/" + interval1 + "/" + interval2 + "/" + interval3 + "/" + sound + "/" + metronome + "/" + tempo + "/" + repeats + "/" + silent
		}

		// Read the selects and returns a proposed filename for the etude to be downloaded.
		function etudeFileName() {
		  key = document.getElementById("key-select").value
		  if (key=="random") {
			  key=randomKey()
			  };
		  scale = document.getElementById("scale-select").value
		  interval1 = document.getElementById("interval1-select").value
		  interval2 = document.getElementById("interval2-select").value
		  interval3 = document.getElementById("interval3-select").value
		  sound = document.getElementById("sound-select").value
		  metronome = document.getElementById("metro-select").value
		  tempo = document.getElementById("tempo-select").value
		  repeats = document.getElementById("repeat-select").value
		  silent = document.getElementById("silence-select").value
		  if (scale=="interval"){
			  return scale + "_" + interval1 + "_" + sound + "_" + metronome + "_" + tempo + "_" + repeats  + "_"+ silent + ".midi" 
		  }
		  if (scale=="intervalpair"){
			  return scale + "_" + interval1 + "_" + interval2 + "_" + sound + "_" + metronome + "_" + tempo + "_" + repeats  + "_" + silent + ".midi" 
		  }
		  if (scale=="intervaltriple"){
			  return scale + "_" + interval1 + "_" + interval2 + "_"  + interval3 + "_" + sound + "_" + metronome + "_" + tempo + "_" + repeats  + "_" + silent + ".midi" 
		  }
		  // any other scale 
		  return key + "_" + scale + "_" + sound + "_" + metronome + "_" + tempo + "_" + repeats  + "_" + silent + ".midi"
		}
		// randomKey returns a keyname chosen randomly from a list of supported
		// keys.
		function randomKey() {
			keys = ['c', 'dflat', 'd', 'eflat', 'e', 'f',
			'gflat', 'g', 'aflat', 'a', 'bflat', 'b']
			return keys[Math.floor(Math.random() * keys.length)]
		}

		function playStart() {
			MIDIjs.stop()
			var url = etudeURL()
			if (url != "") {
			  MIDIjs.play(url)
			}
		}

		function playStop() {
		    MIDIjs.stop()
		}
        
		function downloadEtude() {
          var url = etudeURL()
		  if (url == "") {
			  return // bad selection
		  }
		  // adapted from https://stackoverflow.com/a/49917066/426853
		  let a = document.createElement('a')
		  a.href = url
		  a.download = etudeFileName()
		  document.body.appendChild(a)
		  a.click()
		  document.body.removeChild(a)
		}

		// Run start when the doc is fully loaded.
		document.addEventListener("DOMContentLoaded", start);
	`)
	return
}
