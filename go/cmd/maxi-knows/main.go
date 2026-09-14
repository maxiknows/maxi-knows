package main

import (
    "html/template"
    "log"
    "net/http"
)

// PageData contains the data that can be passed from the Go backend
// to the HTML templates.
type PageData struct {
    Username string
    Flashes  []string
}

// aboutHandler handles GET requests to /about.
// It loads the shared layout and the about page template,
// then renders them together and sends the result to the browser.
func aboutHandler(w http.ResponseWriter, r *http.Request) {

    // Load the templates needed for the about page.
    tmpl, err := template.ParseFiles(
        "templates/layout.html",
        "templates/about.html",
    )

    // Stop and return a 500 error if the templates cannot be loaded.
    if err != nil {
        http.Error(w, "Could not load template", http.StatusInternalServerError)
        return
    }

    // Create the data object that is passed to the templates.
    // Username and Flashes are empty for now.
    data := PageData{}

    // Render the named "layout" template and send the generated HTML
    // to the browser through the ResponseWriter.
    err = tmpl.ExecuteTemplate(w, "layout", data)

    // Return a 500 error if the template could not be rendered.
    if err != nil {
        http.Error(w, "Could not render template", http.StatusInternalServerError)
    }
}

func main() {

    // Route GET requests for /about to the aboutHandler function.
    http.HandleFunc("GET /about", aboutHandler)

    // Serve files from the static folder.
    // For example:
    // /static/style.css -> static/style.css
    // /static/monkgroup.png -> static/monkgroup.png
    http.Handle(
        "/static/",
        http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
    )

    // Start the web server on port 8080.
    log.Println("Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}