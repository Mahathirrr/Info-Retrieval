// main.go
package main

import (
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const ITEMS_PER_PAGE = 10

func main() {
	r := gin.Default()
	r.Static("/static", "./static")
	r.SetFuncMap(templateFunctions())
	r.LoadHTMLGlob("templates/*")

	r.GET("/", indexHandler)
	r.POST("/search", searchHandler)
	r.GET("/search", searchHandlerGet)
	r.POST("/refresh-search-engine", refreshSearchEngine)

	// r.GET("/generate-index", func(c *gin.Context) {
	// 	articles, err := loadArticles()
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 		return
	// 	}
	//
	// 	if err := GenerateInvertedIndexJSON(articles); err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 		return
	// 	}
	//
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "Inverted index has been generated successfully!",
	// 		"file":    "inverted_index.json",
	// 	})
	// })

	r.Run(":8080")
}

// Template functions
func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"iterate": func(start, end int) []int {
			result := make([]int, end-start+1)
			for i := start; i <= end; i++ {
				result[i-start] = i
			}
			return result
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"hasPrefix": strings.HasPrefix,
		"trimURLPath": func(url string) string {
			// Hapus protokol
			url = strings.TrimPrefix(url, "https://")
			url = strings.TrimPrefix(url, "http://")

			parts := strings.SplitN(url, "/", 2)
			if len(parts) > 1 {
				path := parts[1]
				if len(path) > 50 {
					path = path[:47] + "..."
				}
				return path
			}
			return ""
		},
	}
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func searchHandler(c *gin.Context) {
	query := c.PostForm("query")
	method := c.PostForm("method")
	// Redirect ke GET untuk handle pagination
	c.Redirect(http.StatusFound, "/search?q="+query+"&method="+method+"&page=1")
}

func searchHandlerGet(c *gin.Context) {
	query := c.Query("q")
	method := c.Query("method")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	allResults := searching(query, method)
	totalResults := len(allResults)
	totalPages := int(math.Ceil(float64(totalResults) / float64(ITEMS_PER_PAGE)))

	if page < 1 {
		page = 1
	} else if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	var pagedResults []SearchResult
	if totalResults > 0 {
		start := (page - 1) * ITEMS_PER_PAGE
		end := start + ITEMS_PER_PAGE
		if end > totalResults {
			end = totalResults
		}
		pagedResults = allResults[start:end]
	}

	c.HTML(http.StatusOK, "results.html", gin.H{
		"results":      pagedResults,
		"query":        query,
		"method":       method,
		"currentPage":  page,
		"totalPages":   totalPages,
		"totalResults": totalResults,
		"previousPage": page - 1,
		"nextPage":     page + 1,
		"showPrevious": page > 1,
		"showNext":     page < totalPages,
	})
}

// Tambahkan endpoint untuk memperbarui data jika diperlukan
func refreshSearchEngine(c *gin.Context) {
	if err := searchEngine.Refresh(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Search engine data refreshed successfully"})
}
