package config

import (
	"github.com/chromedp/cdproto/network"
	"regexp"
)

var IgnoredMimeTypes = []string{
	"image/vnd.microsoft.icon", // favicon
	"image/png",
}

var IgnoredHostsWithSubdomains = []string{
	"google-analytics.com",
	"maps.googleapis.com",
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
}

var IgnoredNetworkResponseTypes = []network.ResourceType{
	"Stylesheet",
	"Font",
	"Image",
}

var IgnoredPathPatterns = []*regexp.Regexp{
	// Not contains jquery and extension js in file name
	regexp.MustCompile(`(?i)(jquery|\.js)$`),
}
