package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func sampleLetter() Letter {
	return Letter{
		Date: "September 25, 2026", Business: "Juniper Coffee", Contact: "Jordan Lee",
		Address: "123 Main Street", City: "Tulsa, OK 74103",
		Headline:    "Make your business easier to run.",
		Opening:     "Hello, I’m Phillip England. I build polished websites and practical software for small businesses.",
		Opportunity: "A clear, mobile-friendly website can answer customers’ questions and give them a simple way to get in touch.",
		Offer:       "My starter website is $300 for an agreed home page and service pages. Hosting is separate.",
		NextStep:    "Visit englandsoftware.com to start a conversation.",
		Theme:       "color",
	}
}

func TestRenderPDF(t *testing.T) {
	l := sampleLetter()
	if err := validate(&l); err != nil {
		t.Fatal(err)
	}
	colorPDF, err := renderPDF(l)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(colorPDF, []byte("%PDF-")) {
		t.Fatal("output is not a PDF")
	}
	if len(colorPDF) < 10000 {
		t.Fatalf("PDF unexpectedly small: %d bytes", len(colorPDF))
	}
	l.Theme = "mono"
	monoPDF, err := renderPDF(l)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(colorPDF, monoPDF) {
		t.Fatal("black and white setting did not change the PDF")
	}
}

func TestValidationAndOverflow(t *testing.T) {
	l := sampleLetter()
	l.Business = " "
	if err := validate(&l); err == nil {
		t.Fatal("missing business was accepted")
	}
	l = sampleLetter()
	l.Theme = "invalid"
	if err := validate(&l); err == nil {
		t.Fatal("invalid print style was accepted")
	}
	l = sampleLetter()
	l.Headline = strings.Repeat("A long heading ", 6)
	l.Opening = strings.Repeat("A detailed observation about this business. ", 10)
	l.Opportunity = strings.Repeat("This process could be clearer for customers. ", 10)
	l.Offer = strings.Repeat("I could build a straightforward solution for this. ", 9)
	l.NextStep = strings.Repeat("I would be glad to talk this through. ", 8)
	if err := validate(&l); err != nil {
		t.Fatal(err)
	}
	_, err := renderPDF(l)
	if !errors.Is(err, errOverflow) {
		t.Fatalf("wanted page overflow error, got %v", err)
	}
}

func TestBuiltWebsiteLetter(t *testing.T) {
	l := sampleLetter()
	l.Kind = "built"
	l.Headline = "I built a website for your business."
	l.Opening = "I put together a live website concept for your business. Explore it at the link below."
	l.NextStep = "Text or call me at (918) 500-8396 if you would like to purchase it."
	l.Opportunity = ""
	l.Offer = ""
	l.DemoURL = "https://demo.example.com/juniper"
	l.RemovalDate = "2026-11-01"
	if err := validate(&l); err != nil {
		t.Fatal(err)
	}
	pdf, err := renderPDF(l)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(pdf, []byte(l.DemoURL)) {
		t.Fatal("PDF does not link to the live demo")
	}
	if bytes.Contains(pdf, []byte("https://englandsoftware.com")) {
		t.Fatal("built letter QR link points to the company site")
	}
	for _, value := range []string{"", "example.com", "ftp://example.com", "https://user:pass@example.com"} {
		invalid := l
		invalid.DemoURL = value
		if err := validate(&invalid); err == nil {
			t.Fatalf("accepted invalid demo URL %q", value)
		}
	}
	invalid := l
	invalid.RemovalDate = "2026-99-99"
	if err := validate(&invalid); err == nil {
		t.Fatal("accepted invalid removal date")
	}
}
