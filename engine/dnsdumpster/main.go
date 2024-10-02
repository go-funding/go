package dnsdumpster

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/multierr"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const dnsDumpsterURL = "https://dnsdumpster.com"

func getCSRFTokenFromBody(body []byte) (token string, err error) {
	// Define the regular expression pattern
	pattern := `csrfmiddlewaretoken.*?value="([^"]*)`

	// Compile the regular expression
	re := regexp.MustCompile(pattern)

	// Find the first match
	match := re.FindSubmatch(body)

	if len(match) > 1 {
		return string(match[1]), nil
	}
	return "", errors.New("token not found")
}

func getCSRFToken() (token string, err error) {
	// Make a GET request to the URL
	resp, err := http.Get(dnsDumpsterURL)
	if err != nil {
		return
	}
	defer multierr.AppendInvoke(&err, multierr.Invoke(resp.Body.Close))

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return getCSRFTokenFromBody(body)
}

type TxtRecord struct {
	Contents string
}

type DnsServer struct {
	IP               string
	Name             string
	Provider         string
	ProviderLocation string
}

type MxRecord struct {
	IP               string
	Name             string
	Provider         string
	ProviderLocation string
}

type ARecord struct {
	IP               string
	Name             string
	Provider         string
	ProviderLocation string
}

type DnsDumpsterResponse struct {
	Dns []DnsServer
	Txt []TxtRecord
	Mx  []MxRecord
	A   []ARecord
}

func trim(v string) string {
	return strings.Trim(v, " \n\t")
}

func Run(domain string) (response DnsDumpsterResponse, err error) {
	csrfToken, err := getCSRFToken()
	if err != nil {
		return
	}

	// Create a new HTTP client
	client := &http.Client{}

	// Prepare the form data
	form := url.Values{}
	form.Add("csrfmiddlewaretoken", csrfToken)
	form.Add("targetip", domain)
	form.Add("user", "free")

	// Create a new POST request
	req, err := http.NewRequest("POST", dnsDumpsterURL, strings.NewReader(form.Encode()))
	if err != nil {
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", dnsDumpsterURL)
	req.Header.Set("Cookie", "csrftoken="+csrfToken)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer multierr.AppendInvoke(&err, multierr.Invoke(resp.Body.Close))

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table").Each(func(i int, selection *goquery.Selection) {
		headerTxt := selection.Parent().PrevFiltered("p").Text()

		if strings.Contains(headerTxt, "TXT Records") {
			selection.Find("tr").Each(func(i int, sel *goquery.Selection) {
				response.Txt = append(response.Txt, TxtRecord{
					Contents: trim(sel.Find("td").Text()),
				})
			})
			return
		}

		if strings.Contains(headerTxt, "DNS Servers") {
			selection.Find("tr").Each(func(i int, selection *goquery.Selection) {
				providerAll := trim(selection.Find("td:nth-child(3)").Text())
				providerLocation := trim(selection.Find("td:nth-child(3) > span").Text())

				response.Dns = append(response.Dns, DnsServer{
					Name: trim(selection.Find("td:nth-child(1)").
						Nodes[0].FirstChild.Data),
					IP:               trim(selection.Find("td:nth-child(2)").Text()),
					Provider:         providerAll[:strings.Index(providerAll, providerLocation)],
					ProviderLocation: providerLocation,
				})
			})
			return
		}

		if strings.Contains(headerTxt, "MX Records") {
			selection.Find("tr").Each(func(i int, selection *goquery.Selection) {
				providerAll := trim(selection.Find("td:nth-child(3)").Text())
				providerLocation := trim(selection.Find("td:nth-child(3) > span").Text())

				response.Mx = append(response.Mx, MxRecord{
					Name: trim(selection.Find("td:nth-child(1)").
						Nodes[0].FirstChild.Data),
					IP:               trim(selection.Find("td:nth-child(2)").Text()),
					Provider:         providerAll[:strings.Index(providerAll, providerLocation)],
					ProviderLocation: providerLocation,
				})
			})
			return
		}

		if strings.Contains(headerTxt, "Host Records") {
			selection.Find("tr").Each(func(i int, selection *goquery.Selection) {
				providerAll := trim(selection.Find("td:nth-child(3)").Text())
				providerLocation := trim(selection.Find("td:nth-child(3) > span").Text())

				response.A = append(response.A, ARecord{
					Name: trim(selection.Find("td:nth-child(1)").
						Nodes[0].FirstChild.Data),
					IP:               trim(selection.Find("td:nth-child(2)").Text()),
					Provider:         providerAll[:strings.Index(providerAll, providerLocation)],
					ProviderLocation: providerLocation,
				})
			})
			return
		}

		return // not found
	})

	return
}
