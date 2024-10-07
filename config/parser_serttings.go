package config

import (
	"github.com/chromedp/cdproto/network"
	"regexp"
)

var IgnoredMimeTypes = []string{
	"image/vnd.microsoft.icon", // favicon
	"image/png",
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/svg+xml",
	"image/webp",
	"image/tiff",
	"image/bmp",

	"video/mp4",
	"video/mpeg",
	"video/ogg",
	"video/webm",
	"video/x-msvideo",
	"video/quicktime",
	"video/x-matroska",
	"video/3gpp",
	"video/x-flv",
}

var IgnoredHostsWithSubdomains = []string{
	"google-analytics.com",
	"hsadspixel.net",
	"hs-analytics.net",
	"hs-banner.net",
	"hs-scripts.net",
	"hubspot.com",
	"hubapi.com",
	"intercom.io",
	"intercomcdn.com",
	"stripe.com",
	"hscollectedforms.net",
	"hubspotwebflow.net",
	"godaddy.com",
	"g2crowd.com",
	"googleapis.com",
	"fonts.gstatic.com",
	"apis.google.com",
	"googletagmanager.com",
	"www.gstatic.com",
	"g.doubleclick.net",
	"youtube.com",
	"fontawesome.com",
	"connect.facebook.net",
	"sentry.io",
	"cdnjs.cloudflare.com",
	"cdn.jsdelivr.net",
	"cdn.onesignal.com",
	"cdn.rudderlabs.com",
	"cdn.segment.com",
	"cdn.shopify.com",
	"cdn.zapier.com",
	"cdnjs.cloudflare.com",
	"clearbitjs.com",
	"widget-v3.smartsuppcdn.com",
	"widget.trustpilot.com",
	"clarity.ms",
	"unpkg.com",
	"linkedin.com",
	"tiktok.com",
	"forms.hubspot.com",
	"ws.zoominfo.com",
	"events.api.secureserver.net",
	"wixstatic.com",
	"paypal.com",
	"gatsbyjs.com",
	"google.com",
	"google.pt",
	"bing.com",
	"bing.pt",
	"performance.squarespace.com",
	"ipblocker.io",
	"parastorage.com",
	"cookielaw.org",
	"getresponse.pl",
	"getresponse.com",
}

var IgnoredNetworkResponseTypes = []network.ResourceType{
	"Stylesheet",
	"Font",
	"Image",
}

var IgnoredPathPatterns = []*regexp.Regexp{
	// Not contains jquery and extension js in file name
	regexp.MustCompile(`(?i)(jquery([a-zA-Z0-9.-_]*)\.js\?)`),
	regexp.MustCompile(`(?i)(angular([a-zA-Z0-9.-_]*)\.js\?)`),
}
