# Offerit

A local Go application for making personalized, one-page PDF offer letters for England Software. It supports a general introduction and an offer for a website you have already built. Drafts stay in this browser, and each preview or download is generated in memory by the Go server.

## Run

```sh
make run
```

Open [http://127.0.0.1:8080](http://127.0.0.1:8080). Set `OFFERIT_ADDR` to use a different listen address, for example `OFFERIT_ADDR=127.0.0.1:9090 make run`.

## Write a letter

1. Choose **General introduction** to describe your website and software services. Its QR code points to `englandsoftware.com`.
2. Choose **Website already built** to share a live demo. Enter its full `http` or `https` URL; the PDF prints the link and generates a QR code for that address. You can add a demo removal date.
3. Add the business name and tailor the message. The built website letter includes the supplied `assets/cost.png` graphic showing $500 for the website and first year of hosting, or $550 with the optional setup service. It invites the recipient to text or call your phone number to purchase.
4. Choose a print style, check the preview, and download the PDF.

The PDF is US Letter sized and limited to one page. If the copy does not fit, shorten the text until the preview updates. The upper right stays clear for a standard business card. There is no account, database, or external API.

## Check

```sh
go test ./...
```

Fonts are Manrope, Space Grotesk, and Space Mono under the included SIL Open Font Licenses in `assets/`.
