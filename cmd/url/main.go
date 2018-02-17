package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/template"
)

type values map[string]interface{}

type flatURL struct {
	Scheme         string        `json:"scheme"`
	User           *url.Userinfo `json:"user,omitempty"`
	Hostname       string        `json:"hostname"`
	Host           string        `json:"host"`
	Path           string        `json:"path"`
	PathComponents []string      `json:"pathComponents"`
	RawQuery       string        `json:"rawQuery,omitempty"`
	Query          values        `json:"query"`
	Port           string        `json:"port"`
	Fragment       string        `json:"fragment"`
	Opaque         string        `json:"opaque,omitempty"`
	RawPath        string        `json:"-"`
	ForceQuery     bool          `json:"-"`
}

type configuration struct {
	template            string // -template="..."
	resolve             bool   // -resolve
	printVersionAndExit bool   // -version
}

var config configuration
var version string
var outputTemplate *template.Template
var templateFuncs = map[string]interface{}{}

func init() {
	flag.StringVar(&config.template, "template", "", "go template output")
	flag.StringVar(&config.template, "t", "", "alias for -template")
	flag.BoolVar(&config.resolve, "resolve", false, "resolve ../ in URLs")
	flag.BoolVar(&config.resolve, "r", false, "alias for -resolve")
	flag.BoolVar(&config.printVersionAndExit, "version", false, "print version and exit")
	flag.Parse()

	if config.printVersionAndExit {
		fmt.Println(version)
		os.Exit(0)
	}

	if config.template != "" {
		var err error
		if !strings.Contains(config.template, "{{") {
			config.template = "{{" + config.template + "}}"
		}
		outputTemplate, err = template.New("url").Funcs(templateFuncs).Parse(config.template)
		if err != nil {
			fmt.Fprintln(os.Stderr, "template parse error:", err)
			os.Exit(1)
		}
	}
}

func main() {
	exitCode := 0
	enc := json.NewEncoder(os.Stdout)
	for _, urlString := range flag.Args() {
		rawURL, err := url.Parse(urlString)
		if err != nil {
			fmt.Fprintln(os.Stderr, "URL parse error:", err)
			exitCode = 1
			continue
		}
		if rawURL.Scheme == "" {
			opaque := rawURL.Opaque
			rawURL.Scheme = "dummy"
			fixedURL, err := url.Parse(rawURL.String())
			rawURL.Scheme = ""
			if err == nil {
				rawURL = fixedURL
				rawURL.Scheme = ""
				rawURL.Opaque = opaque
			}
		}
		if config.resolve {
			empty := url.URL{}
			rawURL = empty.ResolveReference(rawURL)
		}

		query := values{}
		for k, v := range rawURL.Query() {
			query[k] = v
			if len(v) == 1 {
				query[k] = v[0]
			}
		}

		u := flatURL{
			Scheme:         rawURL.Scheme,
			Opaque:         rawURL.Opaque,
			User:           rawURL.User,
			Host:           rawURL.Host,
			Hostname:       rawURL.Hostname(),
			Path:           rawURL.Path,
			PathComponents: strings.Split(strings.TrimPrefix(rawURL.Path, "/"), "/"),
			RawQuery:       rawURL.RawQuery,
			Query:          query,
			Fragment:       rawURL.Fragment,
			RawPath:        rawURL.RawPath,
			ForceQuery:     rawURL.ForceQuery,
			Port:           rawURL.Port(),
		}

		if outputTemplate != nil {
			outputTemplate.Execute(os.Stdout, u)
			os.Stdout.Write([]byte{'\n'})
		} else {
			enc.Encode(u)
		}
	}
	os.Exit(exitCode)
}
