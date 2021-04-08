package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
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

		tmplMap[basename] = string(contents)
		// fmt.Println(filename, tmpl)
	}
	fmt.Println(tmplMap)
	// return a map
	// return tmplMap

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
	// p, err := loadPage("Home")
	// if err != nil {
	// 	p = &Page{Title: "Home", Body: []byte("No content! Edit this page!")}
	// }
	// renderTemplate(w, "view.html", p)
	html := render("home")
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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
