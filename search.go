package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/RadhiFadlillah/go-sastrawi"
)

// Struktur dasar
type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

type SearchResult struct {
	Title              string
	Content            string
	URL                string
	Score              float64
	HighlightedContent template.HTML
	Favicon            string
}

// Struktur untuk inverted index
type InvertedIndex struct {
	Index map[string]*PostingList
}

type PostingList struct {
	DocFrequency int
	Postings     map[int]*Posting
}

type Posting struct {
	DocID     int
	Frequency int
	Positions []int
}

// Text Processor dengan Sastrawi
type TextProcessor struct {
	stemmer     sastrawi.Stemmer
	stopwords   sastrawi.Dictionary
	punctuation *regexp.Regexp
	numbers     *regexp.Regexp
}

// Tambahkan struktur untuk menyimpan data di memori
type SearchEngine struct {
	Articles      []Article
	InvertedIndex *InvertedIndex
	TFIDFScores   map[string]map[int]float64
	mutex         sync.RWMutex
}

var (
	searchEngine *SearchEngine
	once         sync.Once
)

// Inisialisasi search engine
func initSearchEngine() {
	once.Do(func() {
		searchEngine = &SearchEngine{}
		if err := searchEngine.Initialize(); err != nil {
			log.Fatalf("Failed to initialize search engine: %v", err)
		}
	})
}

// Method untuk inisialisasi SearchEngine
func (se *SearchEngine) Initialize() error {
	// Load articles
	articles, err := loadArticles()
	if err != nil {
		return err
	}

	// Build inverted index
	invertedIndex := buildInvertedIndex(articles)

	// Calculate TF-IDF scores
	tfidfScores := calculateTFIDF(invertedIndex, len(articles))

	// Set data ke struct
	se.mutex.Lock()
	se.Articles = articles
	se.InvertedIndex = invertedIndex
	se.TFIDFScores = tfidfScores
	se.mutex.Unlock()

	return nil
}

// Method untuk refresh data (jika diperlukan)
func (se *SearchEngine) Refresh() error {
	return se.Initialize()
}

func NewTextProcessor() *TextProcessor {
	// Membuat stemmer dan dictionary
	dictionary := sastrawi.DefaultDictionary()
	stemmer := sastrawi.NewStemmer(dictionary)
	stopwords := sastrawi.DefaultStopword()

	return &TextProcessor{
		stemmer:     stemmer,
		stopwords:   stopwords,
		punctuation: regexp.MustCompile(`[^\w\s]`),
		numbers:     regexp.MustCompile(`\b\d+\b`),
	}
}

var textProcessor *TextProcessor

func init() {
	textProcessor = NewTextProcessor()
	initSearchEngine()
}

// Text Processing Steps
// 1. Remove punctuations dan nomor/angka
func (tp *TextProcessor) removePunctuationsAndNumbers(text string) string {
	text = tp.punctuation.ReplaceAllString(text, " ")
	text = tp.numbers.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// 2. Case folding
func (tp *TextProcessor) caseFolding(text string) string {
	return strings.ToLower(text)
}

// 3. Tokenisasi
func (tp *TextProcessor) tokenize(text string) []string {
	return sastrawi.Tokenize(text)
}

// 4. Remove Stopwords dan Stemming dengan Sastrawi
func (tp *TextProcessor) processWords(tokens []string) []string {
	result := make([]string, 0)
	for _, token := range tokens {
		// Skip jika stopword
		if tp.stopwords.Contains(token) {
			continue
		}
		// Stem kata
		stemmed := tp.stemmer.Stem(token)
		if stemmed != "" {
			result = append(result, stemmed)
		}
	}
	return result
}

// Proses text lengkap dengan urutan yang benar
func (tp *TextProcessor) ProcessText(text string) []string {
	// 1. Remove punctuations dan nomor/angka
	cleaned := tp.removePunctuationsAndNumbers(text)

	// 2. Case folding
	folded := tp.caseFolding(cleaned)

	// 3. Tokenisasi
	tokens := tp.tokenize(folded)

	// 4. Remove stopwords dan stemming
	return tp.processWords(tokens)
}

// Fungsi untuk membuat inverted index baru
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		Index: make(map[string]*PostingList),
	}
}

// Fungsi untuk membangun inverted index
func buildInvertedIndex(articles []Article) *InvertedIndex {
	idx := NewInvertedIndex()

	for docID, article := range articles {
		tokens := textProcessor.ProcessText(article.Title + " " + article.Content)

		// Track position untuk setiap term
		for pos, token := range tokens {
			if _, exists := idx.Index[token]; !exists {
				idx.Index[token] = &PostingList{
					DocFrequency: 0,
					Postings:     make(map[int]*Posting),
				}
			}

			if _, exists := idx.Index[token].Postings[docID]; !exists {
				idx.Index[token].Postings[docID] = &Posting{
					DocID:     docID,
					Frequency: 0,
					Positions: make([]int, 0),
				}
				idx.Index[token].DocFrequency++
			}

			posting := idx.Index[token].Postings[docID]
			posting.Frequency++
			posting.Positions = append(posting.Positions, pos)
		}
	}

	return idx
}

// Menghitung TF-IDF dengan inverted index
func calculateTFIDF(invertedIndex *InvertedIndex, totalDocs int) map[string]map[int]float64 {
	tfidfScores := make(map[string]map[int]float64)

	for term, postingList := range invertedIndex.Index {
		tfidfScores[term] = make(map[int]float64)

		idf := math.Log(float64(totalDocs) / float64(postingList.DocFrequency))

		for docID, posting := range postingList.Postings {
			tf := float64(posting.Frequency)
			tfidfScores[term][docID] = tf * idf
		}
	}

	return tfidfScores
}

// Normalisasi vector
func normalizeVector(vector map[string]float64) map[string]float64 {
	normalized := make(map[string]float64)
	var magnitude float64

	for _, weight := range vector {
		magnitude += weight * weight
	}
	magnitude = math.Sqrt(magnitude)

	if magnitude > 0 {
		for term, weight := range vector {
			normalized[term] = weight / magnitude
		}
	}

	return normalized
}

// Cosine Similarity dengan TF-IDF
func cosineSimilarityWithTFIDF(queryVector map[string]float64, tfidfScores map[string]map[int]float64, docID int) float64 {
	docVector := make(map[string]float64)
	for term, scores := range tfidfScores {
		if score, exists := scores[docID]; exists {
			docVector[term] = score
		}
	}

	normalizedQuery := normalizeVector(queryVector)
	normalizedDoc := normalizeVector(docVector)

	var dotProduct, queryMagnitude, docMagnitude float64

	for term, queryWeight := range normalizedQuery {
		queryMagnitude += queryWeight * queryWeight
		if docWeight, exists := normalizedDoc[term]; exists {
			dotProduct += queryWeight * docWeight
		}
	}

	for _, docWeight := range normalizedDoc {
		docMagnitude += docWeight * docWeight
	}

	queryMagnitude = math.Sqrt(queryMagnitude)
	docMagnitude = math.Sqrt(docMagnitude)

	return dotProduct / (queryMagnitude * docMagnitude)
}

// Jaccard Similarity dengan TF-IDF
func jaccardSimilarityWithTFIDF(queryVector map[string]float64, tfidfScores map[string]map[int]float64, docID int) float64 {
	querySet := make(map[string]bool)
	docSet := make(map[string]bool)

	for term := range queryVector {
		querySet[term] = true
	}

	for term, scores := range tfidfScores {
		if _, exists := scores[docID]; exists {
			docSet[term] = true
		}
	}

	intersection := 0
	for term := range querySet {
		if docSet[term] {
			intersection++
		}
	}

	union := len(querySet) + len(docSet) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// Content Preview Generator
func getContentPreview(content, query string, maxLength int) string {
	cleanedContent := cleanContent(content)
	maxLength = 160

	if len(cleanedContent) <= maxLength {
		return cleanedContent
	}

	queryTokens := textProcessor.ProcessText(query)
	contentTokens := textProcessor.ProcessText(cleanedContent)

	queryText := strings.Join(queryTokens, " ")
	contentText := strings.Join(contentTokens, " ")

	pos := strings.Index(strings.ToLower(contentText), strings.ToLower(queryText))
	if pos == -1 {
		return cleanedContent[:maxLength] + "..."
	}

	words := strings.Fields(cleanedContent)
	wordCount := len(strings.Fields(contentText[:pos]))

	wordPos := 0
	for i := 0; i < wordCount && i < len(words); i++ {
		wordPos += len(words[i]) + 1
	}

	start := wordPos - 60
	if start < 0 {
		start = 0
	}

	end := start + maxLength
	if end > len(cleanedContent) {
		end = len(cleanedContent)
	}

	result := cleanedContent[start:end]
	if start > 0 {
		result = "..." + result
	}
	if end < len(cleanedContent) {
		result = result + "..."
	}

	return result
}

// Clean content for better processing
func cleanContent(content string) string {
	unwantedTexts := []string{
		"Baca juga:", "Baca Juga:",
		"Simak breaking news", "Google News",
		"Terus ikuti", "Lebih banyak informasi",
		"Follow", "Instagram", "Twitter", "Facebook",
		"Bagikan:", "Share:", "Read more",
	}

	for _, text := range unwantedTexts {
		content = strings.ReplaceAll(content, text, "")
	}

	urlPatterns := []*regexp.Regexp{
		regexp.MustCompile(`https?://\S+`),
		regexp.MustCompile(`www\.\S+`),
		regexp.MustCompile(`\S+\.(com|net|org)\S*`),
	}

	for _, pattern := range urlPatterns {
		content = pattern.ReplaceAllString(content, "")
	}

	emailPattern := regexp.MustCompile(`\S+@\S+\.\S+`)
	content = emailPattern.ReplaceAllString(content, "")

	socialPattern := regexp.MustCompile(`@\S+`)
	content = socialPattern.ReplaceAllString(content, "")

	specialCharsPattern := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	content = specialCharsPattern.ReplaceAllString(content, " ")

	numberPattern := regexp.MustCompile(`\s+\d+\s+`)
	content = numberPattern.ReplaceAllString(content, " ")

	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
	content = strings.TrimSpace(content)

	content = regexp.MustCompile(`\s*[.,!?;:]\s*`).ReplaceAllString(content, " ")

	words := strings.Fields(content)
	uniqueWords := make([]string, 0)
	prev := ""
	for _, word := range words {
		if word != prev {
			uniqueWords = append(uniqueWords, word)
			prev = word
		}
	}
	content = strings.Join(uniqueWords, " ")

	content = strings.TrimSpace(content)
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	return content
}

// Highlight matched text
func highlightText(text string, query string) string {
	if query == "" {
		return text
	}

	queryTokens := textProcessor.ProcessText(query)
	highlighted := text

	for _, token := range queryTokens {
		if len(token) < 2 {
			continue
		}
		pattern := `(?i)\b[\wа-я]*` + regexp.QuoteMeta(token) + `[\wа-я]*\b`
		re := regexp.MustCompile(pattern)
		highlighted = re.ReplaceAllString(highlighted, `<em>$0</em>`)
	}

	return highlighted
}

// Get favicon path for URL
func getFaviconPath(url string) string {
	switch {
	case strings.HasPrefix(url, "https://artikel.rumah123.com/"):
		return "/static/rumah123.png"
	case strings.HasPrefix(url, "https://propertiterkini.com/"):
		return "/static/propertiterkini.png"
	case strings.HasPrefix(url, "https://propertyandthecity.com/"):
		return "/static/propertyandthecity.png"
	default:
		return "/static/favicon.svg"
	}
}

// Load articles from JSON file
func loadArticles() ([]Article, error) {
	var allArticles []Article

	data, err := ioutil.ReadFile("articles.json")
	if err != nil {
		log.Printf("Error reading articles.json: %v", err)
		return nil, err
	}

	if err := json.Unmarshal(data, &allArticles); err != nil {
		log.Printf("Error parsing JSON from articles.json: %v", err)
		return nil, err
	}

	return allArticles, nil
}

// Main search function
func searching(query string, method string) []SearchResult {
	// Gunakan read lock untuk membaca data
	searchEngine.mutex.RLock()
	defer searchEngine.mutex.RUnlock()

	queryTokens := textProcessor.ProcessText(query)
	queryVector := make(map[string]float64)
	for _, token := range queryTokens {
		queryVector[token]++
	}

	var results []SearchResult

	for i, article := range searchEngine.Articles {
		var score float64
		switch method {
		case "cosine":
			score = cosineSimilarityWithTFIDF(queryVector, searchEngine.TFIDFScores, i)
		case "jaccard":
			score = jaccardSimilarityWithTFIDF(queryVector, searchEngine.TFIDFScores, i)
		default:
			score = cosineSimilarityWithTFIDF(queryVector, searchEngine.TFIDFScores, i)
		}

		if score > 0 {
			contentPreview := getContentPreview(article.Content, query, 160)
			highlightedContent := highlightText(contentPreview, query)

			results = append(results, SearchResult{
				Title:              article.Title,
				Content:            contentPreview,
				URL:                article.URL,
				Score:              score,
				HighlightedContent: template.HTML(highlightedContent),
				Favicon:            getFaviconPath(article.URL),
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
