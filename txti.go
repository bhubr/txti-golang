package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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

/*-----------*
| Templates  |
*------------*/
func loadTemplates() map[string]string {
	// read templates directory
	files, err := ioutil.ReadDir("templates")
	if err != nil {
		log.Fatal(err)
	}
	// make map
	tmplMap := make(map[string]string)
	// parse templates
	for _, file := range files {
		bits := strings.Split(file.Name(), ".")
		basename := bits[0]
		tmplMap[basename] = basename
		fmt.Println(file.Name(), basename)
	}
	fmt.Println(tmplMap)
	// return a map
	return tmplMap
}

/*-----------*
| Handlers   |
*------------*/
func homeHandler(w http.ResponseWriter, r *http.Request) {
	p, err := loadPage("Home")
	if err != nil {
		p = &Page{Title: "Home", Body: []byte("No content! Edit this page!")}
	}
	renderTemplate(w, "view.html", p)
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
