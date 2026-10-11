package generator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HualiNox/cn-service-cidrs/internal/parser"
)

func TestBuildDomainClassical(t *testing.T) {
	const input = "domain:example.cn\ndomain:sub.example.cn\nfull:www.example.cn\nfull:exact.test\nfull:exact.test\nkeyword:china\nregexp:^api[0-9]{1,3}\\.test$\n"
	const want = "DOMAIN-SUFFIX,example.cn\nDOMAIN,exact.test\nDOMAIN-KEYWORD,china\nDOMAIN-REGEX,^api[0-9]{1,3}\\.test$\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(input))
	}))
	defer server.Close()
	sourceFiles := []parser.SourceFile{{
		Directory: "CN/Services",
		Group: parser.SourceGroup{
			Name:    "Example",
			Sources: []parser.Source{{Type: parser.DomainList, Value: server.URL}},
		},
	}}
	root := t.TempDir()
	output, site := filepath.Join(root, "output"), filepath.Join(root, "site")
	sources, err := Build(output, sourceFiles)
	if err != nil {
		t.Fatal(err)
	}
	if err := BuildSite(output, site, sourceFiles, sources); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{output, site} {
		data, err := os.ReadFile(filepath.Join(dir, "tables", "CN-domain-classical.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != want {
			t.Fatalf("%s classical rules = %q, want %q", dir, data, want)
		}
	}
	metadata, err := os.ReadFile(filepath.Join(site, "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	var page pageData
	if err := json.Unmarshal(metadata, &page); err != nil {
		t.Fatal(err)
	}
	if page.DomainClassicalURL != "tables/CN-domain-classical.txt" || page.DomainCount != 4 {
		t.Fatalf("unexpected domain metadata: URL=%q count=%d", page.DomainClassicalURL, page.DomainCount)
	}
	index, err := os.ReadFile(filepath.Join(site, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(index), `href="tables/CN-domain-classical.txt"`) {
		t.Fatal("index is missing the classical download link")
	}
	if err := os.Remove(filepath.Join(site, "tables", "CN-domain-classical.txt")); err != nil {
		t.Fatal(err)
	}
	hash, err := hashTables(filepath.Join(site, "tables"))
	if err != nil {
		t.Fatal(err)
	}
	if hash == page.ContentHash {
		t.Fatal("classical output was not included in the content hash")
	}
}

func TestWriteDomainClassicalEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CN-domain-classical.txt")
	if err := writeDomainClassical(nil, path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("empty domain rules should not create an output file: %v", err)
	}
}
