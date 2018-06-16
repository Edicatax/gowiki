package main

import (
    "log"
    "regexp"
    "html/template"
    "io/ioutil"
    "net/http"
)

// template.Must is a convenience wrapper that panics on a non-nil err value.
// A panic is fine here because if we don't have edit or view we don't have a 
// website.
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

// Regex for path validation.  MustCompile panics if the regex is bad.
var validPath = regexp.MustCompile("^/(edit|view|save)/([a-zA-Z0-9_-]+)$")

type Page struct {
    Title string
    Body []byte
}

// Returns nil on success, error otherwise
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

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindSubstringMatch(r.URL.Path)
    // If the url isn't a valid match for the regex we're doing it wrong
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid page title")
    }
    return m[2], nil // The title is the second matching group
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl + ".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/view/"):]
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/" + title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/edit/"):]
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title,}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/save/"):]
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/" + title, http.StatusFound)

}

func main() {
    http.HandleFunc("/view/", viewHandler)
    http.HandleFunc("/edit/", editHandler)
    http.HandleFunc("/save/", saveHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
