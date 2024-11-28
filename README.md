# Property News Search Engine

A Go-based search engine application specifically designed for property news articles in Indonesia. This project implements advanced search algorithms including TF-IDF with Cosine and Jaccard similarity measures.

## Project Information

- **Kelompok**: 15
- **Anggota**: Muhammad Mahathir (2208107010056)

## Interface Screenshots

### Search Page
[Place your search page screenshot here]

### Results Page
[Place your results page screenshot here]

## Data Sources

The search engine crawls and indexes articles from three major Indonesian property news portals:
- [PropertyAndTheCity.com](https://propertyandthecity.com)
- [PropertiTerkini.com](https://propertiterkini.com)
- [Rumah123.com (Articles)](https://artikel.rumah123.com)

## Scope (Ruang Lingkup)

1. **Tren Nilai Properti**
   - Analisis pergerakan harga properti
   - Faktor-faktor yang mempengaruhi nilai properti
   - Prediksi dan forecast nilai properti

2. **Berita Bursa Perumahan**
   - Perkembangan pasar properti
   - Kebijakan dan regulasi properti
   - Tren dan dinamika pasar perumahan

3. **Listing Properti**
   - Informasi properti yang dijual/disewa
   - Detail dan spesifikasi properti
   - Perbandingan harga properti

## Features

- **Advanced Search Algorithms**
  - TF-IDF (Term Frequency-Inverse Document Frequency)
  - Cosine Similarity
  - Jaccard Similarity

- **Text Processing**
  - Stemming using Sastrawi (Indonesian language)
  - Stopword removal
  - Case folding
  - Punctuation and number removal

- **Search Results**
  - Relevance-based ranking
  - Content preview with query highlighting
  - Pagination support
  - Source website favicon display

## Technical Stack

- **Backend**: Go (Golang)
- **Web Framework**: Gin
- **Template Engine**: Go HTML Templates
- **Text Processing**: Go-Sastrawi
- **Data Storage**: JSON

## Project Structure

```
.
├── articles.json         # Indexed articles data
├── main.go               # Main application entry point
├── search.go             # Search engine implementation
├── static/               # Static assets (images, favicon)
└── templates/            # HTML templates
    ├── 404.html
    ├── document.html
    └── index.html
```

## Key Components

1. **Search Engine**
   - Inverted index construction
   - TF-IDF score calculation
   - Vector space model implementation
   - Multiple similarity measures

2. **Text Processing Pipeline**
   - Text cleaning and normalization
   - Tokenization
   - Stopword removal
   - Stemming for Indonesian language

3. **Web Interface**
   - Clean and responsive design
   - Real-time search results
   - Article preview with highlighted matches
   - Pagination for large result sets

## Running the Application

1. Make sure you have Go installed on your system
2. Clone the repository
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run the application:
   ```bash
   go run main.go
   ```
5. Access the application at `http://localhost:8080`

## Search Methods

The application supports two search methods:

1. **Cosine Similarity (Default)**
   - Measures the cosine of the angle between two vectors
   - Better for documents of different lengths
   - More precise for content similarity

2. **Jaccard Similarity**
   - Measures similarity based on the intersection over union
   - Good for quick similarity approximations
   - Useful for comparing document sets

## Performance Features

- Thread-safe search engine implementation
- Efficient inverted index structure
- Optimized content preview generation
- Fast search response times
- Memory-efficient data structures

## License

This project is part of an academic assignment at Universitas Syiah Kuala.
