package namegenerator

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"
)

// NameGenerator exposes methods to seed and generate names
type NameGenerator interface {
	// SeedData takes a slice of names and a variant indentifier and seeds a Markov chain with the name data
	SeedData(variant string, names []string)

	//GenerateName returns a randomly generated for a particular variant seeded with SeedData
	GenerateName(variant string) (string, error)

	// Variants returns a slice of available name variants
	Variants() []string
}

// New creates a new empty NameGenerator with no seed data
func New() NameGenerator {
	return &generator{chains: make(map[string]chain)}
}

// chain represents a Markov chain
type chain map[string]map[interface{}]int

func (c chain) increment(key string, token interface{}) {
	if v, ok := c[key]; ok {
		v[token]++
	} else {
		c[key] = make(map[interface{}]int)
		c[key][token] = 1
	}
}

func (c chain) scale() {
	tableLen := make(map[interface{}]int)
	for k, v := range c {
		tableLen[k] = 0

		for token, count := range v {
			weighted := math.Floor(math.Pow(float64(count), 1.3))

			v[token] = int(math.Trunc(weighted))
			tableLen[k] += int(math.Trunc(weighted))
		}
	}
	c["table_len"] = tableLen
}

func (c chain) selectLink(key string) interface{} {
	length := c["table_len"][key]
	idx := int(math.Floor(rand.Float64() * float64(length)))

	t := 0

	for token, value := range c[key] {
		t += value
		if idx < t {
			return token
		}
	}
	return "-"
}

func (c chain) markovName() string {
	parts := c.selectLink("parts").(int)
	names := make([]string, 0)

	for i := 0; i < parts; i++ {
		nameLength := c.selectLink("name_len").(int)
		char := c.selectLink("initial").(string)
		name := char
		lastChar := char

		for len(name) < nameLength {
			char = c.selectLink(lastChar).(string)
			name += char
			lastChar = char
		}
		names = append(names, name)
	}
	return strings.Join(names, " ")

}

// GenerateName generates a random name
func (g generator) GenerateName(variant string) (string, error) {
	chain, ok := g.chains[variant]
	if !ok {
		return "", fmt.Errorf("unable to generate name of type %s: no sample data exists", variant)
	}
	return chain.markovName(), nil
}

type generator struct {
	chains map[string]chain
}

func (g *generator) Variants() []string {
	variants := make([]string, len(g.chains))
	i := 0
	for k := range g.chains {
		variants[i] = k
		i++
	}
	return variants
}

func (g *generator) SeedData(label string, names []string) {
	chain := make(chain)

	for _, v := range names {
		splitNames := regexp.MustCompile(`\s+`).Split(v, -1)
		chain.increment("parts", len(splitNames))
		for _, name := range splitNames {
			chain.increment("name_len", len(name))
			chain.increment("initial", string(name[0]))

			lastChar := rune(name[0])

			for _, c := range name[1:] {
				chain.increment(string(lastChar), string(c))
				lastChar = c
			}

		}
	}

	chain.scale()
	g.chains[label] = chain
}
