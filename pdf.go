package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	qrcode "github.com/skip2/go-qrcode"
)

const (
	pageW  = 612.0
	pageH  = 792.0
	margin = 48.0
)

func renderPDF(l Letter) ([]byte, error) {
	p := fpdf.NewCustom(&fpdf.InitType{OrientationStr: "P", UnitStr: "pt", Size: fpdf.SizeType{Wd: pageW, Ht: pageH}})
	p.SetMargins(0, 0, 0)
	p.SetAutoPageBreak(false, 0)
	p.SetTitle("A note for "+l.Business, false)
	p.SetAuthor("England Software", false)
	p.AddUTF8FontFromBytes("Manrope", "", asset("Manrope.ttf"))
	p.AddUTF8FontFromBytes("Space", "", asset("SpaceGrotesk.ttf"))
	p.AddUTF8FontFromBytes("Mono", "", asset("SpaceMono-Regular.ttf"))
	if err := p.Error(); err != nil {
		return nil, err
	}
	p.AddPage()
	logoOptions := fpdf.ImageOptions{ImageType: "PNG"}
	p.RegisterImageOptionsReader("logo", logoOptions, bytes.NewReader(asset("logo.png")))
	// Keep the mark faint and slightly high on the right, behind the letter.
	const watermarkW = 330.0
	const watermarkH = watermarkW * 402 / 351
	p.SetAlpha(0.065, "Normal")
	p.ImageOptions("logo", pageW-watermarkW-28, 151, watermarkW, watermarkH, false, logoOptions, 0, "")
	p.SetAlpha(1, "Normal")

	ink := [3]int{17, 17, 17}
	muted := [3]int{85, 85, 85}
	blue := [3]int{49, 87, 189}
	if l.Theme == "mono" {
		blue = [3]int{35, 35, 35}
	}
	line := [3]int{207, 208, 210}
	color := func(rgb [3]int) { p.SetTextColor(rgb[0], rgb[1], rgb[2]) }
	stroke := func(rgb [3]int) { p.SetDrawColor(rgb[0], rgb[1], rgb[2]) }
	text := func(family string, size, x, y float64, value string, rgb [3]int) {
		p.SetFont(family, "", size)
		color(rgb)
		p.Text(x, y, value)
	}
	paragraph := func(value string, x, y, width, size, leading float64, rgb [3]int) float64 {
		p.SetFont("Manrope", "", size)
		color(rgb)
		p.SetXY(x, y)
		p.MultiCell(width, leading, value, "", "L", false)
		return p.GetY()
	}

	// The upper right is deliberately clear for a 3.5 x 2 inch business card.
	p.ImageOptions("logo", margin, 40, 43, 43, false, logoOptions, 0, "")
	text("Space", 16, 103, 58, "England Software", ink)
	text("Mono", 7.4, 103, 76, "MAKE YOUR BUSINESS EASIER TO RUN.", muted)
	stroke(line)
	p.SetLineWidth(.7)
	p.Line(margin, 106, 306, 106)
	stroke(blue)
	p.SetLineWidth(2.2)
	p.Line(margin, 106, 91, 106)

	p.SetFillColor(blue[0], blue[1], blue[2])
	p.Rect(margin, 144, 3, 13, "F")
	text("Mono", 8, margin+12, 155, strings.ToUpper(l.Date), blue)
	text("Mono", 7.5, margin, 184, "PREPARED FOR", muted)
	y := 201.0
	addressLine := func(value string, size, leading float64, rgb [3]int) {
		p.SetFont("Manrope", "", size)
		for _, line := range p.SplitLines([]byte(value), 250) {
			text("Manrope", size, margin, y, string(line), rgb)
			y += leading
		}
	}
	addressLine(l.Business, 12, 16, ink)
	for _, item := range []string{l.Contact, l.Address, l.Address2, l.City} {
		if item != "" {
			addressLine(item, 9.5, 15, muted)
		}
	}
	if y > 279 {
		return nil, errOverflow
	}

	// Headline uses a generous line height, then the body flows within the fixed footer.
	y = 274
	p.SetFont("Space", "", 28)
	color(ink)
	p.SetXY(margin, y)
	p.MultiCell(pageW-2*margin, 33, l.Headline, "", "L", false)
	y = p.GetY()
	y += 12
	y = paragraph(l.Opening, margin, y, pageW-2*margin, 10.5, 16, ink)
	if l.Kind == "built" {
		y += 17
		if y > 395 {
			return nil, errOverflow
		}
		stroke(line)
		p.SetLineWidth(.6)
		p.Line(margin, y, pageW-margin, y)
		y += 19
		text("Mono", 8, margin, y, "YOUR LIVE WEBSITE", blue)
		y += 7
		y = paragraph(l.DemoURL, margin, y, pageW-2*margin, 9, 13, ink)
		y += 8
		costOptions := fpdf.ImageOptions{ImageType: "PNG"}
		p.RegisterImageOptionsReader("cost", costOptions, bytes.NewReader(asset("cost.png")))
		p.ImageOptions("cost", margin, y, 360, 360*916/1564, false, costOptions, 0, "")
		y += 360*916/1564 + 8
		if y > 651 {
			return nil, errOverflow
		}
		y = paragraph(l.NextStep, margin, y, pageW-2*margin, 10, 15, ink)
		if l.RemovalDate != "" {
			y += 8
			date, _ := time.Parse("2006-01-02", l.RemovalDate)
			text("Mono", 8, margin, y, "DEMO AVAILABLE UNTIL "+strings.ToUpper(date.Format("January 2, 2006")), blue)
		}
		if y > 690 {
			return nil, errOverflow
		}
	} else {
		y += 25
		if y > 477 {
			return nil, errOverflow
		}
		stroke(line)
		p.SetLineWidth(.6)
		p.Line(margin, y, pageW-margin, y)
		y += 20
		text("Mono", 8, margin, y, "01  /  THE OPPORTUNITY", blue)
		y += 12
		y = paragraph(l.Opportunity, margin, y, pageW-2*margin, 10, 15, ink)
		y += 23
		text("Mono", 8, margin, y, "02  /  WHAT I CAN BUILD", blue)
		y += 12
		y = paragraph(l.Offer, margin, y, pageW-2*margin, 10, 15, ink)
		y += 19
		if y > 653 {
			return nil, errOverflow
		}
		y = paragraph(l.NextStep, margin, y, pageW-2*margin, 10, 15, ink)
		if y > 688 {
			return nil, errOverflow
		}
	}

	// The footer is a compact, printable call to action with ample QR quiet zone.
	p.Line(margin, 711, pageW-margin, 711)
	stroke(blue)
	p.SetLineWidth(2.2)
	p.Line(margin, 711, 91, 711)
	qrURL := "https://englandsoftware.com"
	if l.Kind == "built" {
		qrURL = l.DemoURL
	}
	qrPNG, err := qrcode.Encode(qrURL, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("QR code: %w", err)
	}
	p.RegisterImageOptionsReader("qr", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(qrPNG))
	p.ImageOptions("qr", pageW-margin-58, 722, 58, 58, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	text("Manrope", 12, margin, 739, "Phillip England", ink)
	text("Manrope", 8.5, margin, 754, "Founder, England Software", muted)
	footerURL := "ENGLANDSOFTWARE.COM"
	if l.Kind == "built" {
		footerURL = "SCAN TO VIEW THE LIVE DEMO"
	}
	text("Mono", 8, margin, 773, footerURL, blue)
	email := l.SenderEmail
	if email == "" {
		email = "phillip@englandsoftware.com"
	}
	if l.Kind == "built" {
		text("Mono", 7.5, 234, 739, "TEXT OR CALL TO PURCHASE", blue)
		text("Manrope", 10, 234, 756, "(918) 500-8396", ink)
	} else {
		text("Manrope", 8.5, 234, 754, email, ink)
		text("Manrope", 8.5, 234, 773, "(918) 500-8396", muted)
	}
	p.LinkString(pageW-margin-58, 722, 58, 58, qrURL)
	if err := p.Error(); err != nil {
		return nil, fmt.Errorf("PDF layout: %w", err)
	}
	var out bytes.Buffer
	if err := p.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
