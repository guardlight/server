package analysismanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileNameBuilding(t *testing.T) {
	filename := makeFilePath("/books/", "Fantasy", stripLeadingArticle("Book name by the author"))
	filenameWithoutArticle := makeFilePath("/books/", "Science Fiction", stripLeadingArticle("The book name by the author"))
	filenameOther := makeFilePath("/books/", "romance", stripLeadingArticle("Kone book name by the author"))
	filenameOtherEmpty := makeFilePath("/books/", "romance", stripLeadingArticle("K book name by the author"))
	filenameOtherNonAlpha := makeFilePath("/books/", "romance", stripLeadingArticle("K;book name by the author"))

	assert.Equal(t, "/books/fantasy/bo/book_name_by_the_author.txt", filename)
	assert.Equal(t, "/books/science_fiction/bo/book_name_by_the_author.txt", filenameWithoutArticle)
	assert.Equal(t, "/books/romance/ko/kone_book_name_by_the_author.txt", filenameOther)
	assert.Equal(t, "/books/romance/k_/k_book_name_by_the_author.txt", filenameOtherEmpty)
	assert.Equal(t, "/books/romance/kb/kbook_name_by_the_author.txt", filenameOtherNonAlpha)

	// Short title (1 letter)
	filenameShort := makeFilePath("/books/", "Drama", stripLeadingArticle("Z"))
	assert.Equal(t, "/books/drama/z_/z.txt", filenameShort)

	// Empty title
	filenameEmpty := makeFilePath("/books/", "Mystery", stripLeadingArticle(""))
	assert.Equal(t, "/books/mystery/__/unnamed.txt", filenameEmpty) // your implementation may vary here

	// Title with number
	filenameWithNumber := makeFilePath("/books/", "Horror", stripLeadingArticle("1984"))
	assert.Equal(t, "/books/horror/19/1984.txt", filenameWithNumber)

	// Title with symbol
	filenameWithSymbol := makeFilePath("/books/", "Philosophy", stripLeadingArticle("#Hashtag Life"))
	assert.Equal(t, "/books/philosophy/ha/hashtag_life.txt", filenameWithSymbol)

	// Title with mixed case and spaces
	filenameMixed := makeFilePath("/books/", "Sci-Fi", stripLeadingArticle("An Interesting Case"))
	assert.Equal(t, "/books/sci_fi/in/interesting_case.txt", filenameMixed)

	// Unicode characters (should normalize or strip depending on your implementation)
	filenameUnicode := makeFilePath("/books/", "World Literature", stripLeadingArticle("Étranger"))
	assert.Equal(t, "/books/world_literature/ét/étranger.txt", filenameUnicode) // depends on slug logic

	// Title with multiple articles
	filenameMultipleArticles := makeFilePath("/books/", "Action", stripLeadingArticle("The An A Case"))
	assert.Equal(t, "/books/action/an/an_a_case.txt", filenameMultipleArticles) // only removes first article

	// Title with tab or newline
	filenameWeirdWhitespace := makeFilePath("/books/", "Poetry", stripLeadingArticle("The\nSilent\tVoice"))
	assert.Equal(t, "/books/poetry/si/silent_voice.txt", filenameWeirdWhitespace)

	// Title with dashes, underscores, and punctuation
	filenameSpecialChars := makeFilePath("/books/", "Thriller", stripLeadingArticle("The Book: Name! (2020)"))
	assert.Equal(t, "/books/thriller/bo/book_name_2020.txt", filenameSpecialChars)
}

func TestWriteToFile(t *testing.T) {
	err := writeToTextToFile("/docker-compose/guardlight/hello-test-file.txt", "hello")
	assert.NoError(t, err)
}
