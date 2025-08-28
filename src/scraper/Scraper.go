package scraper

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func ScrapeVagalume(query string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Encode query to avoid issues with special characters
	encodedQuery := url.QueryEscape(query)
	fmt.Println("Step 1: Encoding query and initializing context")

	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-webgl", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("no-first-run", true),
		chromedp.UserAgent(GetRandomUserAgent()),
	}

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, options...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Navigate to search page
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(fmt.Sprintf("https://www.vagalume.com.br/search?q=%s", encodedQuery)),
		chromedp.WaitReady("body", chromedp.ByQuery),
	)
	if err != nil {
		return "", fmt.Errorf("failed to navigate to search page: %v", err)
	}

	// Check for "Nenhum resultado"
	var searchContent string
	err = chromedp.Run(taskCtx,
		chromedp.Text("body", &searchContent, chromedp.ByQuery),
	)
	if err != nil {
		return "", fmt.Errorf("failed to extract search page content: %v", err)
	}
	if strings.Contains(searchContent, "Nenhum resultado") {
		return "", fmt.Errorf("no lyrics found for the search: %s", query)
	}

	fmt.Println("Step 2: Search page loaded, no 'Nenhum resultado' found")

	var href string
	err = chromedp.Run(taskCtx,
		chromedp.AttributeValue("a.gs-title", "href", &href, nil, chromedp.ByQuery),
	)
	if err != nil {
		return "", fmt.Errorf("failed to extract href: %v", err)
	}

	fmt.Println("Step 3: Extracted href:", href)

	// Remove "-traducao" from the href
	if strings.Contains(href, "-traducao") {
		href = strings.ReplaceAll(href, "-traducao", "")
	}
	fmt.Println("Step 4: Processed href (removed -traducao if present):", href)

	// Navigate to the extracted href with timeout
	navCtx, cancelNav := context.WithTimeout(taskCtx, 10*time.Second)
	defer cancelNav()

	err = chromedp.Run(navCtx,
		chromedp.Navigate(href),
		chromedp.WaitReady("body", chromedp.ByQuery),
	)
	if err != nil {
		return "", fmt.Errorf("failed to navigate to href: %v", err)
	}

	fmt.Println("Step 5: Navigated to href:", href)

	// Check if the page has explicit content by looking at hasBadwords in vData
	var vData string
	err = chromedp.Run(taskCtx,
		chromedp.Text(`script#vData`, &vData, chromedp.ByQuery),
	)
	if err != nil {
		fmt.Printf("Step 6: Failed to extract vData, proceeding with modal check: %v\n", err)
	} else if !strings.Contains(vData, `hasBadwords:1`) {
		fmt.Println("Step 6: No explicit content (hasBadwords:0), skipping modal check")
	} else {
		// Page has explicit content, check for age restriction modal
		const modalSelector = "button.ageRestrictionModal-button.stylePrimary.h18"
		modalCtx, cancelModal := context.WithTimeout(taskCtx, 2*time.Second)
		defer cancelModal()

		err = chromedp.Run(modalCtx,
			chromedp.WaitVisible(modalSelector, chromedp.ByQuery),
		)
		if err == nil {
			// Modal button is visible, attempt to click it
			err = chromedp.Run(taskCtx,
				chromedp.Click(modalSelector, chromedp.ByQuery, chromedp.NodeVisible),
			)
			if err != nil {
				return "", fmt.Errorf("failed to click age button: %v", err)
			}
			fmt.Println("Step 6: Clicked age restriction button")
		} else {
			// Modal not found, try fallback selector
			const fallbackSelector = "button[class*='ageRestrictionModal-button']"
			fallbackCtx, cancelFallback := context.WithTimeout(taskCtx, 1*time.Second)
			defer cancelFallback()
			err = chromedp.Run(fallbackCtx,
				chromedp.WaitVisible(fallbackSelector, chromedp.ByQuery),
			)
			if err == nil {
				err = chromedp.Run(taskCtx,
					chromedp.Click(fallbackSelector, chromedp.ByQuery, chromedp.NodeVisible),
				)
				if err != nil {
					return "", fmt.Errorf("failed to click fallback age button: %v", err)
				}
				fmt.Println("Step 6: Clicked fallback age restriction button")
			} else {
				// Log minimal HTML for debugging
				var bodyHTML string
				_ = chromedp.Run(taskCtx,
					chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery),
				)
				fmt.Printf("Step 6: Age restriction modal button (%s) and fallback (%s) not found, proceeding without click. Error: %v\nBody HTML (truncated): %s\n", modalSelector, fallbackSelector, err, bodyHTML[:500])
			}
		}
	}

	var lyrics string
	err = chromedp.Run(taskCtx,
		chromedp.Text("div#lyrics", &lyrics, chromedp.ByQuery),
	)
	if err != nil {
		return "", fmt.Errorf("failed to extract lyrics: %v", err)
	}

	fmt.Println("Step 7: Extracted lyrics content")

	// Remove the age confirmation text
	ageConfirmationText := `Confirmação de Idade

Este conteúdo é destinado a maiores de 18 anos, por conter material impróprio para menores.

Ao prosseguir, você declara que tem 18 anos ou mais e está autorizado a acessá-lo.

SOU MAIOR DE 18 ANOS`
	lyrics = strings.ReplaceAll(lyrics, ageConfirmationText, "")

	fmt.Println("Step 8: Cleaned lyrics content")

	return lyrics, nil
}
