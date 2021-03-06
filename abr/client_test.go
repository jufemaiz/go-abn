package abr

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/h2non/gock"
)

const TEST_ABR_GUID = "TEST_ABR_GUID"

func init() {
	os.Setenv("ABR_GUID", TEST_ABR_GUID)
}

func TestSimple(t *testing.T) {
	defer gock.Off()

	gock.New("http://foo.com").
		Get("/bar").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	res, err := http.Get("http://foo.com/bar")
	if err != nil {
		t.Errorf("Expected %v, got %v", nil, err)
	}
	if res.StatusCode != 200 {
		t.Errorf("Expected %v, got %v", 200, res.StatusCode)
	}

	body, _ := ioutil.ReadAll(res.Body)
	if string(body)[:13] != `{"foo":"bar"}` {
		t.Errorf("Expected %v, got %v", `{"foo":"bar"}`, string(body)[:13])
	}

	// Verify that we don't have pending mocks
	if !gock.IsDone() {
		t.Errorf("Expected %v, got %v", true, gock.IsDone())
	}
}

func TestABRClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Error(err)
		return
	}

	if client.BaseURL.String() != BaseURL {
		t.Errorf("Expected endpoint to be %s, got %s", BaseURL, client.BaseURL.String())
	}
}

func TestABRClientNoEnvSet(t *testing.T) {
	guid := os.Getenv("ABR_GUID")
	os.Unsetenv("ABR_GUID")
	defer os.Setenv("ABR_GUID", guid)

	_, err := NewClient()
	if err == nil {
		t.Errorf("Expected an error, none was raised")
	} else if err.Error() != MissingGUIDError {
		t.Error(err)
	}

	return
}

var abnTestCases = []struct {
	abn      string
	name     string
	filename string
}{
	{"99124391073", "COzero Pty Ltd", "abn/200/99124391073.xml"},
	{"26154482283", "Oneflare Pty Ltd", "abn/200/26154482283.xml"},
	{"65433405893", "STUART J AULD", "abn/200/65433405893.xml"},
}

var asicTestCases = []struct {
	abn      string
	acn      string
	name     string
	filename string
}{
	{"78159033075", "159033075", "ENERGYLINK GLOBAL PTY LTD", "acn/200/159033075.xml"},
	{"26154482283", "154482283", "Oneflare Pty Ltd", "acn/200/154482283.xml"},
}

func TestSearchByABNv201408(t *testing.T) {
	defer gock.Off()

	client, err := NewClient()
	if err != nil {
		t.Error(err)
		return
	}

	for _, c := range abnTestCases {
		body, err := ioutil.ReadFile(filepath.Join("testdata", c.filename))
		reqBody := url.Values{}
		reqBody.Set("authenticationGuid", TEST_ABR_GUID)
		reqBody.Add("includeHistoricalDetails", "Y")
		reqBody.Add("searchString", c.abn)

		gock.New("https://www.abn.business.gov.au/").
			Post("/abrxmlsearch/ABRXMLSearch.asmx/SearchByABNv201408").
			MatchType("url").
			BodyString(reqBody.Encode()).
			Reply(200).
			BodyString(string(body))

		entity, err := client.SearchByABNv201408(c.abn, true)
		if err != nil {
			t.Error(err)
			continue
		}

		if entity.Name() != c.name {
			t.Errorf("Expected %v, got %v", c.name, entity.Name())
		}

		if entity.ABN() != c.abn {
			t.Errorf("Expected %v, got %v", c.abn, entity.ABN())
		}
	}
	return
}

func TestSearchByASICv201408(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Error(err)
		return
	}

	for _, c := range asicTestCases {
		body, err := ioutil.ReadFile(filepath.Join("testdata", c.filename))
		reqBody := url.Values{}
		reqBody.Set("authenticationGuid", TEST_ABR_GUID)
		reqBody.Add("includeHistoricalDetails", "Y")
		reqBody.Add("searchString", c.acn)

		gock.New("https://www.abn.business.gov.au/").
			Post("/abrxmlsearch/ABRXMLSearch.asmx/SearchByASICv201408").
			MatchType("url").
			BodyString(reqBody.Encode()).
			Reply(200).
			BodyString(string(body))

		entity, err := client.SearchByASICv201408(c.acn, true)
		if err != nil {
			t.Error(err)
			continue
		}

		if entity.Name() != c.name {
			t.Errorf("Expected %v, got %v", c.name, entity.Name())
		}

		if entity.ABN() != c.abn {
			t.Errorf("Expected %v, got %v", c.abn, entity.ABN())
		}

		if entity.ASICNumber != c.acn {
			t.Errorf("Expected %v, got %v", c.acn, entity.ASICNumber)
		}
	}
	return
}
