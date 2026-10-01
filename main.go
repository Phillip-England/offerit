package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed web/* assets/*
var files embed.FS

type Letter struct {
	Kind        string `json:"kind"`
	Date        string `json:"date"`
	Contact     string `json:"contact"`
	Business    string `json:"business"`
	Address     string `json:"address"`
	Address2    string `json:"address2"`
	City        string `json:"city"`
	Headline    string `json:"headline"`
	Opening     string `json:"opening"`
	Opportunity string `json:"opportunity"`
	Offer       string `json:"offer"`
	NextStep    string `json:"nextStep"`
	SenderEmail string `json:"senderEmail"`
	DemoURL     string `json:"demoUrl"`
	RemovalDate string `json:"removalDate"`
	Theme       string `json:"theme"`
}

func main() {
	tmpl := template.Must(template.ParseFS(files, "web/index.html"))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, map[string]string{"Today": time.Now().Format("January 2, 2006")})
	})
	mux.Handle("GET /assets/", http.FileServerFS(files))
	mux.Handle("GET /web/", http.FileServerFS(files))
	mux.HandleFunc("POST /api/pdf", pdfHandler)

	addr := os.Getenv("OFFERIT_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	log.Printf("Offerit is ready at http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func pdfHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var letter Letter
	if err := json.NewDecoder(r.Body).Decode(&letter); err != nil {
		http.Error(w, "Could not read the letter details.", http.StatusBadRequest)
		return
	}
	if err := validate(&letter); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := renderPDF(letter)
	if err != nil {
		log.Printf("render PDF: %v", err)
		http.Error(w, "Could not render this letter. Try shorter text in the letter sections.", http.StatusUnprocessableEntity)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	if r.URL.Query().Get("download") == "1" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName(letter.Business)))
	} else {
		w.Header().Set("Content-Disposition", "inline")
	}
	_, _ = w.Write(data)
}

func validate(l *Letter) error {
	if l.Kind == "" {
		l.Kind = "general"
	}
	if l.Kind != "general" && l.Kind != "built" {
		return errors.New("Letter type is invalid.")
	}
	limits := []struct {
		name     string
		value    *string
		max      int
		required bool
	}{
		{"Date", &l.Date, 60, true},
		{"Contact", &l.Contact, 100, false},
		{"Business name", &l.Business, 100, true},
		{"Street address", &l.Address, 120, false},
		{"Address line 2", &l.Address2, 120, false},
		{"City, state and ZIP", &l.City, 120, false},
		{"Headline", &l.Headline, 100, true},
		{"Opening", &l.Opening, 450, true},
		{"Opportunity", &l.Opportunity, 500, l.Kind == "general"},
		{"Offer", &l.Offer, 500, l.Kind == "general"},
		{"Next step", &l.NextStep, 350, true},
		{"Your email", &l.SenderEmail, 120, false},
		{"Live demo URL", &l.DemoURL, 500, l.Kind == "built"},
		{"Removal date", &l.RemovalDate, 60, false},
	}
	if l.Kind == "built" {
		u, err := url.Parse(l.DemoURL)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil {
			return errors.New("Enter a valid http or https live demo URL.")
		}
		if strings.ContainsAny(l.DemoURL, "\r\n\t ") {
			return errors.New("Live demo URL cannot contain spaces.")
		}
		if l.RemovalDate != "" {
			if _, err := time.Parse("2006-01-02", l.RemovalDate); err != nil {
				return errors.New("Choose a valid demo removal date.")
			}
		}
	}
	for _, field := range limits {
		*field.value = strings.TrimSpace(*field.value)
		if field.required && *field.value == "" {
			return fmt.Errorf("%s is required.", field.name)
		}
		if len([]rune(*field.value)) > field.max {
			return fmt.Errorf("%s is too long (maximum %d characters).", field.name, field.max)
		}
	}
	if l.Theme == "" {
		l.Theme = "color"
	}
	if l.Theme != "color" && l.Theme != "mono" {
		return errors.New("Print style is invalid.")
	}
	return nil
}

func fileName(name string) string {
	var out strings.Builder
	for _, r := range strings.ToLower(name) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			out.WriteRune(r)
		} else if out.Len() > 0 && !strings.HasSuffix(out.String(), "-") {
			out.WriteByte('-')
		}
	}
	slug := strings.Trim(out.String(), "-")
	if slug == "" {
		slug = "business"
	}
	return "england-software-offer-" + slug + ".pdf"
}

func asset(name string) []byte {
	b, err := files.ReadFile("assets/" + name)
	if err != nil {
		panic(err)
	}
	return b
}

// Kept separate so renderPDF can return an ordinary error on layout overflow.
var errOverflow = errors.New("letter content is too long for one page")
