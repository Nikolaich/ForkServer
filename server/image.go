package server

import (
	"net/http"
	"strconv"
	"strings"
)

var images = map[byte]string{
	'f': "M11 5c-1.629 0-2.305-1.058-4-3h-7v20h24v-17h-13z",
	'l': "M4 22h-4v-4h4v4zm0-12h-4v4h4v-4zm0-8h-4v4h4v-4zm3 0v4h17v-4h-17zm0 12h17v-4h-17v4zm0 8h17v-4h-17v4z",
	't': "M8 24l3-9h-9l14-15-3 9h9l-14 15z",
	'd': "M17 13h6l-11 11-11-11h6v-13h10z",
	'o': "M 12, 12 m -12, 0 a 12,12 0 1,0 24,0 a 12,12 0 1,0 -24,0",
	'a': "M9 18h-7v-12h7v12zm2-12v12l11 6v-24l-11 6z",
	'v': "M16 16c0 1.104-.896 2-2 2h-12c-1.104 0-2-.896-2-2v-8c0-1.104.896-2 2-2h12c1.104 0 2 .896 2 2v8zm8-10l-6 4.223v3.554l6 4.223v-12z",
	's': "M2 7h-2v-2h2c3.49 0 5.48 1.221 6.822 2.854-.41.654-.754 1.312-1.055 1.939-1.087-1.643-2.633-2.793-5.767-2.793zm16 10c-3.084 0-4.604-1.147-5.679-2.786-.302.627-.647 1.284-1.06 1.937 1.327 1.629 3.291 2.849 6.739 2.849v3l6-4-6-4v3zm0-10v3l6-4-6-4v3c-5.834 0-7.436 3.482-8.85 6.556-1.343 2.921-2.504 5.444-7.15 5.444h-2v2h2c5.928 0 7.543-3.511 8.968-6.609 1.331-2.893 2.479-5.391 7.032-5.391z",
}

func image(w http.ResponseWriter, p string) {
	l, c := len(p), "white"
	if !strings.HasSuffix(p, ".svg") {
		panic(404)
	} else if p = strings.TrimSuffix(p, ".svg"); l == 0 {
		panic(404)
	} else if _, o := images[p[0]]; !o {
		panic(404)
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte("<?xml version=\"1.0\" standalone=\"no\"?>\n"))
	w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="`))
	w.Write([]byte(images[p[0]]))
	if l > 1 {
		for _, i := range p[1:] {
			if _, o := images[byte(i)]; o {
				w.Write([]byte(`" fill="` + c + `"/><path transform="matrix(0.6 0 0 0.6 5 5)" d="` + images[byte(i)]))
				c = "black"
			} else if i == 'g' {
				c = "limegreen"
			} else if i == 'r' {
				c = "salmon"
			} else if i == 'y' {
				c = "yellow"
			} else if g, _ := strconv.Atoi(string(i)); g > 0 && g < 4 {
				w.Write([]byte(`" transform="rotate(` + strconv.Itoa(g*90) + ` 12 12)`))
			}
		}
	}
	w.Write([]byte(`" fill="` + c + `"/></svg>`))
}
