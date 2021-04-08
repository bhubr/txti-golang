package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"path"
	"strings"

	"github.com/aymerick/raymond"
)

type Page struct {
	Title string
	Body  []byte
}

/*----------*
| Pages     |
*-----------*/
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

/*-----------*
| Templates  |
*------------*/

// make templates map
var tmplMap = make(map[string]string)
var layoutTmpl *raymond.Template

// load templates from directory
func loadTemplates() {
	// read templates directory
	files, err := ioutil.ReadDir("templates")
	exitOnError(err)

	// read and parse templates
	for _, file := range files {
		// extract basename
		filename := file.Name()
		bits := strings.Split(filename, ".")
		basename := bits[0]

		// read file content
		contents, err := ioutil.ReadFile(path.Join("templates", filename))
		exitOnError(err)

		// parse layout template
		if basename == "layout" {
			layoutTmpl, err = raymond.Parse(string(contents))
			exitOnError(err)
			continue
		}

		// store raw hbs/html template
		tmplMap[basename] = string(contents)
	}

	raymond.RegisterPartials(tmplMap)
}

func render(tmplKey string) string {
	ctx := map[string]interface{}{
		"whichPartial": func() string {
			return tmplKey
		},
	}
	result, err := layoutTmpl.Exec(ctx)
	exitOnError(err)
	return result
}

/*-----------*
| Handlers   |
*------------*/
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmplKey := "home"
	// Fallback to 404
	if r.URL.Path != "/" {
		tmplKey = "404"
	}

	// p, err := loadPage("Home")
	// if err != nil {
	// 	p = &Page{Title: "Home", Body: []byte("No content! Edit this page!")}
	// }
	// renderTemplate(w, "view.html", p)
	html := render(tmplKey)
	fmt.Fprintf(w, html)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	pageSlug := r.URL.Path[len("/view/"):]
	p, err := loadPage(pageSlug)
	// Should not be handled as a regular page
	if err != nil {
		p = &Page{Title: "404 Error", Body: []byte("The page " + pageSlug + " could not be found")}
	}
	renderTemplate(w, "view.html", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	pageSlug := r.URL.Path[len("/edit/"):]
	p, err := loadPage(pageSlug)
	if err != nil {
		p = &Page{Title: pageSlug}
	}
	renderTemplate(w, "edit.html", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	pageSlug := r.URL.Path[len("/save/"):]
	pageBody := r.FormValue("body")
	p := &Page{Title: pageSlug, Body: []byte(pageBody)}
	p.save()
	http.Redirect(w, r, "/view/"+pageSlug, http.StatusFound)
}

// https://golang.org/pkg/crypto/rand/#Read
func generateSlug() string {
	var slug string
	c := 6
	bya := make([]byte, c)
	_, err := rand.Read(bya)
	exitOnError((err))
	// The slice should now contain random bytes instead of only zeroes.
	// fmt.Println(bytes.Equal(b, make([]byte, c)))
	for _, b := range bya {
		// 255 / 7.28 = 35.02275 -> rounds to 35
		ch := int(math.Floor(float64(b) / 7.28))
		base := 48
		if ch >= 10 {
			base = 87
		}
		slug += string(string(rune(base + ch)))
	}
	fmt.Printf("slug: %s\n", slug)
	return slug
}

func createTxtiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// If it's not, use the w.WriteHeader() method to send a 405 status
		// code and the w.Write() method to write a "Method Not Allowed"
		// response body. We then return from the function so that the
		// subsequent code is not executed.
		w.WriteHeader(405)
		w.Write([]byte("Method Not Allowed"))
		return
	}
	// check spambot trap field
	// TODO: log its IP
	spamField := r.FormValue("username")
	if spamField != "" {
		w.WriteHeader(422)
		w.Write([]byte("Unprocessable Entity"))
		return
	}
	// generate slug
	pageSlug := generateSlug()
	// get content
	content := r.FormValue("content")
	p := &Page{Title: pageSlug, Body: []byte(content)}
	p.save()
	http.Redirect(w, r, "/view/"+pageSlug, http.StatusFound)
	// fmt.Fprintf(w, "OK")
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func main() {
	loadTemplates()
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/txtis/create", createTxtiHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
