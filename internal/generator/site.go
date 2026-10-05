package generator

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/HualiNox/cn-service-cidrs/internal/parser"
)

type pageData struct {
	GeneratedAt string      `json:"generated_at"`
	IPv4Count   int         `json:"ipv4_count"`
	IPv6Count   int         `json:"ipv6_count"`
	AllURL      string      `json:"all_url"`
	IPv4URL     string      `json:"ipv4_url"`
	IPv6URL     string      `json:"ipv6_url"`
	Groups      []pageTable `json:"groups"`
	Directories []pageTable `json:"directories"`
}

type pageTable struct {
	Name      string `json:"name"`
	Directory string `json:"directory,omitempty"`
	IPv4Count int    `json:"ipv4_count"`
	IPv6Count int    `json:"ipv6_count"`
	AllURL    string `json:"all_url"`
	IPv4URL   string `json:"ipv4_url,omitempty"`
	IPv6URL   string `json:"ipv6_url,omitempty"`
}

const indexTemplate = `<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="description" content="IPv4 and IPv6 CIDR tables for China services and networks.">
	<title>CN Service CIDRs</title>
	<style>
		:root {
			color-scheme: light dark;
			font: 16px/1.55 system-ui, sans-serif;
		}

		body {
			max-width: 960px;
			margin: 0 auto;
			padding: 2rem 1.25rem 4rem;
		}

		h1,
		h2 {
			line-height: 1.2;
		}

		.muted {
			color: #777;
		}

		.cards {
			display: flex;
			flex-wrap: wrap;
			gap: 1rem;
			margin: 1.5rem 0;
		}

		.card {
			min-width: 9rem;
			border: 1px solid #8885;
			border-radius: .6rem;
			padding: 1rem 1.25rem;
		}

		.card strong {
			display: block;
			font-size: 1.6rem;
		}

		table {
			width: 100%;
			border-collapse: collapse;
			margin: 1rem 0 2rem;
		}

		th,
		td {
			text-align: left;
			border-bottom: 1px solid #8885;
			padding: .65rem .5rem;
		}

		th {
			font-size: .9rem;
		}

		a {
			margin-right: .55rem;
		}

		@media (max-width: 600px) {
			body {
				padding: 1.25rem .75rem 3rem;
			}

			th,
			td {
				padding: .5rem .25rem;
			}
		}
	</style>
</head>
<body>
	<main>
		<h1>CN Service CIDRs</h1>
		<p class="muted">Generated {{.GeneratedAt}} UTC</p>
		<div class="cards">
			<div class="card">
				<span>IPv4 prefixes</span>
				<strong>{{.IPv4Count}}</strong>
			</div>
			<div class="card">
				<span>IPv6 prefixes</span>
				<strong>{{.IPv6Count}}</strong>
			</div>
		</div>

		<h2>China aggregate</h2>
		<p>
			<a href="{{.AllURL}}">All prefixes</a>
			<a href="{{.IPv4URL}}">IPv4</a>
			<a href="{{.IPv6URL}}">IPv6</a>
		</p>

		<h2>Source groups</h2>
		<table>
			<thead>
				<tr>
					<th>Group</th>
					<th>IPv4</th>
					<th>IPv6</th>
					<th>Downloads</th>
				</tr>
			</thead>
			<tbody>
			{{range .Groups}}
			<tr>
				<td>{{.Directory}}/{{.Name}}</td>
				<td>{{.IPv4Count}}</td>
				<td>{{.IPv6Count}}</td>
				<td>
					<a href="{{.AllURL}}">All</a>
					{{if .IPv4URL}}<a href="{{.IPv4URL}}">IPv4</a>{{end}}
					{{if .IPv6URL}}<a href="{{.IPv6URL}}">IPv6</a>{{end}}
				</td>
			</tr>
			{{end}}
			</tbody>
		</table>

		<h2>Directory aggregates</h2>
		<table>
			<thead>
				<tr>
					<th>Directory</th>
					<th>IPv4</th>
					<th>IPv6</th>
					<th>Downloads</th>
				</tr>
			</thead>
			<tbody>
			{{range .Directories}}
			<tr>
				<td>{{.Name}}</td>
				<td>{{.IPv4Count}}</td>
				<td>{{.IPv6Count}}</td>
				<td>
					<a href="{{.AllURL}}">All</a>
					{{if .IPv4URL}}<a href="{{.IPv4URL}}">IPv4</a>{{end}}
					{{if .IPv6URL}}<a href="{{.IPv6URL}}">IPv6</a>{{end}}
				</td>
			</tr>
			{{end}}
			</tbody>
		</table>

		<p class="muted">Machine-readable data: <a href="metadata.json">metadata.json</a></p>
	</main>
</body>
</html>
`

func BuildSite(output, site string, sourceFiles []parser.SourceFile) error {
	if err := mkdirAll(site, true); err != nil {
		return fmt.Errorf("prepare site directory %q: %w", site, err)
	}

	tables := filepath.Join(output, "tables")
	siteTables := filepath.Join(site, "tables")
	if err := copyTables(tables, siteTables); err != nil {
		return fmt.Errorf("copy tables into site: %w", err)
	}

	page, err := newPageData(siteTables, sourceFiles)
	if err != nil {
		return err
	}
	return writeSiteFiles(site, page)
}

func copyTables(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk tables path %q: %w", path, err)
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("get relative table path for %q: %w", path, err)
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("create site directory %q: %w", target, err)
			}
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read table %q: %w", path, err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return fmt.Errorf("copy table to %q: %w", target, err)
		}
		return nil
	})
}

func newPageData(tables string, sourceFiles []parser.SourceFile) (pageData, error) {
	page := pageData{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Groups:      make([]pageTable, 0, len(sourceFiles)),
	}

	for _, sourceFile := range sourceFiles {
		relative := filepath.Join(sourceFile.Directory, sourceFile.Group.Name+".txt")
		group, exists, err := readPageTable(tables, relative, sourceFile.Group.Name, filepath.ToSlash(sourceFile.Directory))
		if err != nil {
			return pageData{}, err
		}
		if exists {
			page.Groups = append(page.Groups, group)
		}
	}

	directories := make(map[string]struct{})
	for _, sourceFile := range sourceFiles {
		for dir := sourceFile.Directory; dir != "."; dir = filepath.Dir(dir) {
			directories[dir] = struct{}{}
		}
	}
	for dir := range directories {
		name := filepath.Base(dir)
		relative := filepath.Join(filepath.Dir(dir), name+".txt")
		summary, exists, err := readPageTable(tables, relative, filepath.ToSlash(dir), "")
		if err != nil {
			return pageData{}, err
		}
		if exists {
			page.Directories = append(page.Directories, summary)
		}
	}
	sort.Slice(page.Directories, func(i, j int) bool {
		return page.Directories[i].Name < page.Directories[j].Name
	})

	china, _, err := readPageTable(tables, "CN.txt", "CN", "")
	if err != nil {
		return pageData{}, err
	}
	page.IPv4Count = china.IPv4Count
	page.IPv6Count = china.IPv6Count
	page.AllURL = china.AllURL
	page.IPv4URL = china.IPv4URL
	page.IPv6URL = china.IPv6URL
	return page, nil
}

func readPageTable(tables, relative, name, directory string) (pageTable, bool, error) {
	allPath := filepath.Join(tables, relative)
	if _, err := os.Stat(allPath); err != nil {
		if os.IsNotExist(err) {
			return pageTable{}, false, nil
		}
		return pageTable{}, false, fmt.Errorf("inspect table %q: %w", allPath, err)
	}

	base := strings.TrimSuffix(relative, ".txt")
	ipv4Count, err := countLines(filepath.Join(tables, base+"-ipv4.txt"))
	if err != nil {
		return pageTable{}, false, err
	}
	ipv6Count, err := countLines(filepath.Join(tables, base+"-ipv6.txt"))
	if err != nil {
		return pageTable{}, false, err
	}

	return pageTable{
		Name:      name,
		Directory: directory,
		IPv4Count: ipv4Count,
		IPv6Count: ipv6Count,
		AllURL:    tableURL(relative),
		IPv4URL:   optionalTableURL(base+"-ipv4.txt", ipv4Count),
		IPv6URL:   optionalTableURL(base+"-ipv6.txt", ipv6Count),
	}, true, nil
}

func countLines(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read table %q: %w", path, err)
	}
	if len(data) == 0 {
		return 0, nil
	}
	count := strings.Count(string(data), "\n")
	if !strings.HasSuffix(string(data), "\n") {
		count++
	}
	return count, nil
}

func optionalTableURL(path string, count int) string {
	if count == 0 {
		return ""
	}
	return tableURL(path)
}

func tableURL(path string) string {
	return "tables/" + filepath.ToSlash(path)
}

func writeSiteFiles(site string, page pageData) error {
	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		return fmt.Errorf("parse index template: %w", err)
	}

	indexPath := filepath.Join(site, "index.html")
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("create index %q: %w", indexPath, err)
	}
	if err := tmpl.Execute(indexFile, page); err != nil {
		_ = indexFile.Close()
		return fmt.Errorf("write index %q: %w", indexPath, err)
	}
	if err := indexFile.Close(); err != nil {
		return fmt.Errorf("close index %q: %w", indexPath, err)
	}

	metadata, err := json.MarshalIndent(page, "", "  ")
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	metadataPath := filepath.Join(site, "metadata.json")
	if err := os.WriteFile(metadataPath, append(metadata, '\n'), 0o644); err != nil {
		return fmt.Errorf("write metadata %q: %w", metadataPath, err)
	}
	return nil
}
