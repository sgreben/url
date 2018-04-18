package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/template"
)

type values map[string]interface{}

type Userinfo struct {
	Username    string
	UsernameSet bool
	Password    string
	PasswordSet bool
}

func (i *Userinfo) fromURLUserinfo(u *url.Userinfo) {
	i.Username = u.Username()
	i.Password, i.PasswordSet = u.Password()
}

func (i *Userinfo) toURLUserinfo() *url.Userinfo {
	if i.PasswordSet {
		return url.UserPassword(i.Username, i.Password)
	}
	if i.UsernameSet {
		return url.User(i.Username)
	}
	return nil
}

type flatURL struct {
	Scheme         string   `json:"scheme"`
	User           Userinfo `json:"user,omitempty"`
	Hostname       string   `json:"hostname"`
	Host           string   `json:"host"`
	Path           string   `json:"path"`
	PathComponents []string `json:"pathComponents"`
	RawQuery       string   `json:"rawQuery,omitempty"`
	Query          values   `json:"query"`
	Port           string   `json:"port"`
	Fragment       string   `json:"fragment"`
	Opaque         string   `json:"opaque,omitempty"`
	RawPath        string   `json:"-"`
	ForceQuery     bool     `json:"-"`
}

type setField struct {
	field string
	value string
}

type stringOrNil struct {
	value    *string
	template *template.Template
}

func (s *stringOrNil) String() string {
	if s.value != nil {
		return string(*s.value)
	}
	return ""
}

func (s *stringOrNil) Set(v string) error {
	if strings.HasPrefix(v, ".") && !strings.Contains(v, "{{") {
		v = "{{" + v + "}}"
	}
	s.value = &v
	t, err := template.New("set").Funcs(templateFuncs).Parse(v)
	s.template = t
	return err
}

type setFields struct {
	setScheme     stringOrNil
	setHostname   stringOrNil
	setUsername   stringOrNil
	setNoUsername bool
	setPassword   stringOrNil
	setNoPassword bool
	setHost       stringOrNil
	setPath       stringOrNil
	setRawQuery   stringOrNil
	setPort       stringOrNil
	setFragment   stringOrNil
	setOpaque     stringOrNil
}

type configuration struct {
	plain               bool
	template            string // -template="..."
	resolve             bool   // -resolve
	get                 string
	setFields                // -set-*="..."
	printVersionAndExit bool // -version
}

var config configuration
var version string
var outputTemplate *template.Template
var templateFuncs = map[string]interface{}{}
var newline = []byte{'\n'}

func init() {
	flag.BoolVar(&config.plain, "plain", false, "plain URL output (useful with -set-* flags)")
	flag.BoolVar(&config.plain, "p", false, "alias for -plain")
	flag.StringVar(&config.template, "template", "", "go template output")
	flag.StringVar(&config.template, "t", "", "alias for -template")
	flag.BoolVar(&config.resolve, "resolve", false, "resolve ../ in URLs")
	flag.BoolVar(&config.resolve, "r", false, "alias for -resolve")
	flag.Var(&config.setScheme, "set-scheme", "set the scheme component")
	flag.Var(&config.setHost, "set-host", "set the host component")
	flag.Var(&config.setUsername, "set-username", "set the username component")
	flag.BoolVar(&config.setNoUsername, "set-no-username", false, "set the username component")
	flag.Var(&config.setPassword, "set-password", "set the password component")
	flag.BoolVar(&config.setNoPassword, "set-no-password", false, "set the password component")
	flag.Var(&config.setPath, "set-path", "set the path component")
	flag.Var(&config.setHostname, "set-hostname", "set the hostname component")
	flag.Var(&config.setPort, "set-port", "set the port component")
	flag.Var(&config.setRawQuery, "set-query", "set the (raw) query component")
	flag.Var(&config.setFragment, "set-fragment", "set the fragment component")
	flag.Var(&config.setOpaque, "set-opaque", "set the opaque component")

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
	b := bytes.NewBuffer(nil)
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
		if config.setScheme.value != nil {
			b.Reset()
			config.setScheme.template.Execute(b, rawURL)
			rawURL.Scheme = b.String()
		}
		if config.setHost.value != nil {
			b.Reset()
			config.setHost.template.Execute(b, rawURL)
			rawURL.Host = b.String()
		}
		if config.setHostname.value != nil {
			port := rawURL.Port()
			b.Reset()
			config.setHostname.template.Execute(b, rawURL)
			rawURL.Host = b.String()
			if port != "" {
				rawURL.Host += ":" + port
			}
		}
		if config.setPort.value != nil {
			hostname := rawURL.Hostname()
			b.Reset()
			config.setPort.template.Execute(b, rawURL)
			rawURL.Host = ":" + b.String()
			if hostname != "" {
				rawURL.Host = hostname + rawURL.Host
			}
		}
		if config.setPath.value != nil {
			b.Reset()
			config.setPath.template.Execute(b, rawURL)
			rawURL.Path = b.String()
		}
		if config.setRawQuery.value != nil {
			b.Reset()
			config.setRawQuery.template.Execute(b, rawURL)
			rawURL.RawQuery = b.String()
		}
		if config.setFragment.value != nil {
			b.Reset()
			config.setFragment.template.Execute(b, rawURL)
			rawURL.Fragment = b.String()
		}
		if config.setOpaque.value != nil {
			b.Reset()
			config.setOpaque.template.Execute(b, rawURL)
			rawURL.Opaque = b.String()
		}

		query := values{}
		for k, v := range rawURL.Query() {
			query[k] = v
			if len(v) == 1 {
				query[k] = v[0]
			}
		}

		user := Userinfo{}
		user.fromURLUserinfo(rawURL.User)
		u := flatURL{
			Scheme:         rawURL.Scheme,
			Opaque:         rawURL.Opaque,
			User:           user,
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

		if config.setUsername.value != nil {
			b.Reset()
			config.setUsername.template.Execute(b, rawURL)
			u.User.Username = b.String()
			u.User.UsernameSet = true
		}
		if config.setNoUsername {
			u.User.UsernameSet = false
		}
		if config.setPassword.value != nil {
			b.Reset()
			config.setPassword.template.Execute(b, rawURL)
			u.User.Password = b.String()
			u.User.PasswordSet = true
		}
		if config.setNoPassword {
			u.User.PasswordSet = false
		}
		rawURL.User = u.User.toURLUserinfo()

		if config.plain {
			os.Stdout.Write([]byte(rawURL.String()))
			os.Stdout.Write(newline)
		} else if outputTemplate != nil {
			outputTemplate.Execute(os.Stdout, u)
			os.Stdout.Write(newline)
		} else {
			enc.Encode(u)
		}
	}
	os.Exit(exitCode)
}
