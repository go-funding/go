package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

func main() {
	// Create a new Chrome instance
	allocCtx, cancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.Flag("auto-open-devtools-for-tabs", true),
	)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Create a WaitGroup to wait for all tabs to be processed
	var wg sync.WaitGroup

	// Set up a listener for tab-related events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *target.EventTargetCreated:
			// New tab created
			if e.TargetInfo.Type == "page" {
				wg.Add(1)
				go handleTab(ctx, e.TargetInfo.TargetID, &wg)
			}
		}
	})

	// Navigate to the download page and click the download button
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.google.com/"),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Browser opened. You can now interact with the page and open new tabs.")
	log.Println("Close a tab when you're ready to download its resources.")

	// Wait for all tabs to be processed
	wg.Wait()

	log.Println("All tabs processed. Exiting.")
}

func downloadAllSources(ctx context.Context, targetID target.ID) error {
	sources, err := page.GetResourceTree().Do(ctx)
	if err != nil {
		return err
	}

	for _, resource := range sources.Resources {
		content, err := page.GetResourceContent(sources.Frame.ID, resource.URL).Do(ctx)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", resource.URL, err)
			continue
		}

		filename := fmt.Sprintf("%s_%s.%s", targetID, resource.URL[len(resource.URL)-10:], resource.Type)
		err = ioutil.WriteFile(filename, content, 0644)
		if err != nil {
			fmt.Printf("Error saving %s: %v\n", filename, err)
			continue
		}

		fmt.Printf("Downloaded: %s\n", filename)
	}

	log.Printf("All resources downloaded for tab %s.\n", targetID)
	return nil
}

func handleTab(parentCtx context.Context, targetID target.ID, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := chromedp.NewContext(parentCtx, chromedp.WithTargetID(targetID))
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := downloadAllSources(ctx, targetID); err != nil {
					log.Printf("Error downloading resources for tab %s: %v\n", targetID, err)
				}
			case <-ctx.Done():
				done <- true
				return
			}
		}
	}()

	<-done
	log.Printf("Tab %s closed.\n", targetID)
}
