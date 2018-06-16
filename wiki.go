package main

import (
    "log"
    "regexp"
    "errors"
    "html/template"
    "io/ioutil"
    "net/http"
)

const templatePath  = "tmpl/"
const dataPath      = "data/"

// template.Must is a convenience wrapper that panics on a non-nil err value.
// A panic is fine here because if we don't have edit or view we don't have a 
// website.
var templates = template.Must(template.ParseFiles(templatePath + "edit.html", templatePath + "view.html"))

// Regex for path validation.  MustCompile panics if the regex is bad.
var validPath = regexp.MustCompile("^/(edit|view|save)/([a-zA-Z0-9_-]+)$")

type Page struct {
    Title string
    Body []byte
}

// Returns nil on success, error otherwise
func (p *Page) save() error {
    filename := dataPath + p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    filename := dataPath + title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

// Deprecated in favour of makeHandler's closure solution
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    // If the url isn't a valid match for the regex we're doing it wrong
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid page title")
    }
    return m[2], nil // The title is the second matching group
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    // ExecuteTemplate wants a file name but not a full path
    err := templates.ExecuteTemplate(w, tmpl + ".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// http.HandleFunc takes an argument of type http.HandlerFunc
func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    /// Closures!
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/" + title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title,}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/" + title, http.StatusFound)

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w,r, "/view/frontpage", http.StatusFound)
}

func main() {
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/", rootHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
