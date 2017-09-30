package isbn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

var (
	apiBase = mustURL("http://isbndb.com/api/v2/json/")

	pkgClient = &http.Client{
		Timeout:   time.Second * 30,
		Transport: http.DefaultTransport,
	}
)

func SetClient(c *http.Client) {
	pkgClient = c
}

func New(key string) (*API, error) {
	b, err := apiBase.Parse(key + "/")
	if err != nil {
		return nil, err
	}
	return &API{
		client: pkgClient,
		base:   b,
	}, nil
}

type API struct {
	client *http.Client
	base   *url.URL
}

type apiError struct {
	E string `json:"error"`
}

func (a *apiError) Error() string {
	return a.E
}

func (a *API) call(ctx context.Context, u *url.URL) (*bytes.Buffer, error) {
	req := (&http.Request{
		Method: http.MethodGet,
		URL:    u,
		Header: http.Header{
			"Accept": {"application/json; charset=utf-8"},
		},
	}).WithContext(ctx)
	res, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %v:%d", u, res.StatusCode)
	}

	buf := &bytes.Buffer{}
	rd := io.TeeReader(res.Body, buf)
	var ae apiError
	if err := json.NewDecoder(rd).Decode(&ae); err != nil {
		return nil, err
	}
	if ae.E != "" {
		return nil, &ae
	}
	return buf, nil
}

func (a *API) Book(ctx context.Context, isbn string) (*Book, error) {
	ret := apiBook{}
	u, err := a.base.Parse(path.Join("book", isbn))
	if err != nil {
		return nil, err
	}
	b, err := a.call(ctx, u)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(b).Decode(&ret); err != nil {
		return nil, err
	}

	return &ret.Books[0], nil
}

func (a *API) Author(ctx context.Context, id string) (*Author, error) {
	ret := apiAuthor{}
	u, err := a.base.Parse(path.Join("author", id))
	if err != nil {
		return nil, err
	}
	b, err := a.call(ctx, u)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(b).Decode(&ret); err != nil {
		return nil, err
	}

	return &ret.Authors[0], nil
}

type apiBook struct {
	Index string `json:"index_searched"`
	Books []Book `json:"data"`
}

type Book struct {
	TitleLatin              string       `json:"title_latin"`
	PhysicalDescriptionText string       `json:"physical_description_text"`
	Summary                 string       `json:"summary"`
	Language                string       `json:"language"`
	SubjectID               []string     `json:"subject_ids"`
	Title                   string       `json:"title"`
	ISBN10                  string       `json:"isbn10"`
	EditionInfo             string       `json:"edition_info"`
	DeweyDecimal            string       `json:"dewey_decimal"`
	MarcEncLevel            string       `json:"marc_enc_level"`
	Notes                   string       `json:"notes"`
	AwardsText              string       `json:"awards_text"`
	DeweyNormal             string       `json:"dewey_normal"`
	URLsText                string       `json:"urls_text"`
	BookID                  string       `json:"book_id"`
	TitleLong               string       `json:"title_long"`
	AuthorData              []BookAuthor `json:"author_data"`
	ISBN13                  string       `json:"isbn13"`
	PublisherText           string       `json:"publisher_text"`
	PublisherName           string       `json:"publisher_name"`
	LCCNumber               string       `json:"lcc_number"`
	PublisherID             string       `json:"publisher_id"`
}

type BookAuthor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type apiAuthor struct {
	Index   string   `json:"index_searched"`
	Authors []Author `json:"data"`
}

type Author struct {
	AuthorID   string   `json:"author_id"`
	CategoryID []string `json:"category_ids"`
	Dates      string   `json:"dates"`
	Name       string   `json:"name"`
	SubjectID  []string `json:"subject_ids"`
	BookCount  string   `json:"book_count"`
	BookID     []string `json:"book_ids"`
	LastName   string   `json:"last_name"`
	NameLatin  string   `json:"name_latin"`
	FirstName  string   `json:"first_name"`
}
