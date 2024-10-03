package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

type tabInfo struct {
	targetID target.ID
	ctx      context.Context
	cancel   context.CancelFunc
}

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("auto-open-devtools-for-tabs", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var wg sync.WaitGroup
	tabMap := make(map[target.ID]*tabInfo)
	var mapMutex sync.Mutex

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *target.EventTargetCreated:
			if e.TargetInfo.Type == "page" {
				tabCtx, tabCancel := chromedp.NewContext(ctx, chromedp.WithTargetID(e.TargetInfo.TargetID))
				mapMutex.Lock()
				tabMap[e.TargetInfo.TargetID] = &tabInfo{
					targetID: e.TargetInfo.TargetID,
					ctx:      tabCtx,
					cancel:   tabCancel,
				}
				mapMutex.Unlock()
				wg.Add(1)
				go handleTab(tabCtx, e.TargetInfo.TargetID, &wg, &mapMutex, tabMap)
			}
		case *target.EventTargetDestroyed:
			mapMutex.Lock()
			if tab, exists := tabMap[e.TargetID]; exists {
				tab.cancel()
				delete(tabMap, e.TargetID)
			}
			mapMutex.Unlock()
		}
	})

	if err := chromedp.Run(ctx, chromedp.Navigate("https://google.com")); err != nil {
		log.Fatal(err)
	}

	log.Println("Browser opened. You can now interact with the page and open new tabs.")
	log.Println("Resources will be downloaded continuously. Close tabs when you're done.")

	wg.Wait()

	log.Println("All tabs processed. Exiting.")
}

func handleTab(ctx context.Context, targetID target.ID, wg *sync.WaitGroup, mapMutex *sync.Mutex, tabMap map[target.ID]*tabInfo) {
	defer wg.Done()

	downloadTicker := time.NewTicker(5 * time.Second)
	defer downloadTicker.Stop()

	for {
		select {
		case <-downloadTicker.C:
			mapMutex.Lock()
			tab, exists := tabMap[targetID]
			mapMutex.Unlock()
			if !exists {
				return
			}
			if err := downloadAllSources(tab.ctx, targetID); err != nil {
				if err == context.Canceled {
					log.Printf("Tab %s closed, stopping downloads.\n", targetID)
					return
				}
				log.Printf("Error downloading resources for tab %s: %v\n", targetID, err)
			}
		case <-ctx.Done():
			log.Printf("Tab %s closed.\n", targetID)
			return
		}
	}
}

func downloadAllSources(ctx context.Context, targetID target.ID) error {
	sources, err := page.GetResourceTree().Do(ctx)
	if err != nil {
		return err
	}

	for _, resource := range sources.Resources {
		if resource.Type == network.ResourceTypeDocument ||
			resource.Type == network.ResourceTypeScript ||
			resource.Type == network.ResourceTypeStylesheet {
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
	}

	return nil
}
