// Package main implementasi web scraper untuk artikel properti
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// Article merepresentasikan struktur data artikel yang akan di-scrape
// Menggunakan tag json untuk format output ke file JSON
type Article struct {
	Title   string `json:"title"`   // Judul artikel
	Content string `json:"content"` // Isi artikel
	URL     string `json:"url"`     // URL sumber artikel
}

// Konstanta warna untuk output terminal
// Menggunakan ANSI escape codes untuk memberikan warna pada output log
const (
	colorRed    = "\033[31m" // Untuk error messages
	colorGreen  = "\033[32m" // Untuk success messages
	colorYellow = "\033[33m" // Untuk warnings
	colorBlue   = "\033[34m" // Untuk informational messages
	colorReset  = "\033[0m"  // Reset warna ke default
)

func main() {
	// Inisialisasi collector dengan konfigurasi:
	// - AllowedDomains: Hanya scrape dari domain yang ditentukan
	// - MaxDepth: Maksimal kedalaman navigasi link
	// - Async: Mengaktifkan scraping asynchronous
	c := colly.NewCollector(
		colly.AllowedDomains("artikel.rumah123.com"),
		colly.MaxDepth(3),
		colly.Async(true),
	)

	// Slice untuk menyimpan semua artikel yang berhasil di-scrape
	var articles []Article

	// Konfigurasi rate limiting untuk menghindari overloading server
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",             // Berlaku untuk semua domain
		RandomDelay: 2 * time.Second, // Delay random antara request
		Parallelism: 4,               // Maksimal 4 request paralel
	})

	// Handler untuk menemukan dan mengunjungi semua link
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		// Hanya proses link dari domain yang ditentukan
		if strings.HasPrefix(link, "https://artikel.rumah123.com/") {
			fmt.Printf("%s[LINK] Found: %s%s\n", colorBlue, link, colorReset)
			e.Request.Visit(link)
		}
	})

	// Handler untuk mengekstrak data artikel
	c.OnHTML("article", func(e *colly.HTMLElement) {
		article := Article{}

		// Ekstrak judul dari elemen h1 dengan class heading-3
		article.Title = strings.TrimSpace(e.ChildText("h1.heading-3"))

		// Ekstrak dan gabungkan konten dari semua tag p dalam div.content
		var contentParts []string
		e.ForEach("div.content p", func(_ int, el *colly.HTMLElement) {
			if text := strings.TrimSpace(el.Text); text != "" {
				contentParts = append(contentParts, text)
			}
		})
		// Gabungkan semua bagian konten dengan newline
		article.Content = strings.Join(contentParts, "\n")

		// Ambil URL dari request saat ini
		article.URL = e.Request.URL.String()

		// Hanya simpan artikel yang memiliki judul dan konten
		if article.Title != "" && article.Content != "" {
			fmt.Printf("%s[ARTICLE] Successfully scraped: %s%s\n",
				colorGreen, article.Title, colorReset)
			articles = append(articles, article)
		}
	})

	// Handler untuk menangani error saat scraping
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("%s[ERROR] Failed to scrape %s: %s%s\n",
			colorRed, r.Request.URL, err, colorReset)
	})

	// Handler yang dijalankan sebelum melakukan request
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("%s[VISITING] %s%s\n",
			colorBlue, r.URL.String(), colorReset)
	})

	// Mulai proses scraping
	fmt.Println("🚀 Starting scraping process...")
	startTime := time.Now()

	// Mulai dari halaman utama
	err := c.Visit("https://artikel.rumah123.com/")
	if err != nil {
		log.Fatal("Failed to start scraping:", err)
	}

	// Tunggu semua goroutine scraping selesai
	c.Wait()

	// Simpan hasil ke file JSON
	outputFile, err := os.Create("articles3.json")
	if err != nil {
		log.Fatal("Failed to create output file:", err)
	}
	defer outputFile.Close()

	// Setup JSON encoder dengan indentasi untuk mudah dibaca
	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")

	// Encode dan tulis ke file
	if err := encoder.Encode(articles); err != nil {
		log.Fatal("Failed to encode articles to JSON:", err)
	}

	// Hitung durasi dan tampilkan statistik
	duration := time.Since(startTime)
	fmt.Printf("\n✨ Scraping completed in %s\n", duration)
	fmt.Printf("📦 Total articles scraped: %d\n", len(articles))
	fmt.Printf("💾 Results saved to articles.json\n")
}
