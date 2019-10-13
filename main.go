package main

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/satori/go.uuid"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))
	http.Handle("/favicon.ico", http.NotFoundHandler())

	p := ":9000"
	fmt.Printf("Server Listening on port %s...\n", p)

	http.ListenAndServe(p, nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	sc := getSessionCookie(w, req)
	fc := getFileCookie(w, req)

	// handle form post
	if req.Method == http.MethodPost {
		nf, fh, err := req.FormFile("nf")
		if err != nil {
			fmt.Println(err)
		}
		defer nf.Close()

		// create sha for filename
		ext := strings.Split(fh.Filename, ".")[1]
		hash := sha1.New()
		io.Copy(hash, nf)
		fname := fmt.Sprintf("%x.%s", hash.Sum(nil), ext)

		// create new file
		dir, err := os.Getwd() // get current working directory
		if err != nil {
			fmt.Println(err)
		}
		path := filepath.Join(dir, "public", "pics", fname)
		f, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
		}
		defer f.Close()

		// copy
		nf.Seek(0, 0)
		io.Copy(f, nf)

		// append new file to cookie
		appendToFileCookie(w, fc, fname)
	}

	// build data for template
	data := struct {
		Session string
		Files   []string
	}{
		Session: sc.Value,
		Files:   strings.Split(fc.Value, "|"),
	}

	tpl.ExecuteTemplate(w, "index.gohtml", data)
}

func getSessionCookie(w http.ResponseWriter, req *http.Request) *http.Cookie {
	c, err := req.Cookie("session")
	if err != nil {
		sID, err := uuid.NewV4()
		if err != nil {
			log.Panicln(err)
		}

		c = &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
	}
	return c
}

func getFileCookie(w http.ResponseWriter, req *http.Request) *http.Cookie {
	c, err := req.Cookie("files")
	if err != nil {
		c = &http.Cookie{
			Name:  "files",
			Value: "empty",
		}
		http.SetCookie(w, c)
	}
	return c
}

func appendToFileCookie(w http.ResponseWriter, c *http.Cookie, fname string) *http.Cookie {
	s := c.Value
	if !strings.Contains(s, fname) {
		s += "|" + fname
	}
	c.Value = s
	http.SetCookie(w, c)
	return c
}
