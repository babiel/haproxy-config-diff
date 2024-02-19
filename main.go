// haproxy-config-diff parset 2 HAProxy-Configs (grob),
// erstellt daraus jeweils eine Datenstruktur
// und vergleicht dann die 2 Datenstrukturen.
//
// Dadurch kann man inhaltliche Änderungen zwischen zwei Configs sehen,
// aber Änderungen an Formatierungen und (irrelevanter) Reihenfolge werden ignoriert.
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
)

type config map[string]map[string]section
type section map[string][]string

func parseConfig(r io.Reader) (*config, error) {
	cfg := make(config)
	var currentSection section

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if s == "" {
			continue
		}

		if strings.HasPrefix(s, "#") {
			continue
		}

		keyword, params, _ := strings.Cut(s, " ")

		switch keyword {
		case "global", "defaults", "listen", "frontend", "backend", "cache", "userlist", "peers":
			// vorherige Section abschließen.
			sort.Strings(currentSection["timeout"])
			sort.Strings(currentSection["acl"])
			sort.Strings(currentSection["bind"])
			sort.Strings(currentSection["server"])
			sort.Strings(currentSection["option"])
			sort.Strings(currentSection["stats"])

			currentSection = make(section)
			if cfg[keyword] == nil {
				cfg[keyword] = make(map[string]section)
			}
			cfg[keyword][params] = currentSection
		default:
			if currentSection == nil {
				return nil, fmt.Errorf("keyword %q must be in a section", keyword)
			}
			currentSection[keyword] = append(currentSection[keyword], params)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func parseConfigFile(fp string) (*config, error) {
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg, err := parseConfig(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", fp, err)
	}

	return cfg, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: haproxy-config-diff FILE FILE")
		os.Exit(1)
	}

	lhs, err := parseConfigFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	rhs, err := parseConfigFile(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	diff := cmp.Diff(lhs, rhs)
	fmt.Println(diff)
}
