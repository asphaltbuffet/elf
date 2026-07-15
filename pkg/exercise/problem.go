package exercise

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"

	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/protocol"
)

// placeholderTitle stands in for a Problem's title when projecteuler.net is
// unreachable, so an offline `add euler` still scaffolds a solvable exercise.
// It is deliberately distinct from any real title so a human can grep for it.
const placeholderTitle = "Untitled"

// declaredParts returns the Parts this Exercise actually has, by Kind. A Project
// Euler Problem has a single answer, so it declares only Part One; an AoC Puzzle
// declares both. The solve/test/benchmark driver iterates this set rather than
// assuming two parts, so a Problem never runs a phantom Part Two.
func (e *Exercise) declaredParts() []protocol.Part {
	if e.Kind == KindProblem {
		return []protocol.Part{protocol.PartOne}
	}

	return []protocol.Part{protocol.PartOne, protocol.PartTwo}
}

// makeProblemID builds the identity string for a Project Euler Problem. Unlike
// an AoC ID (year-day), it is a single unpadded number prefixed with the kind.
func makeProblemID(number int) string {
	return fmt.Sprintf("euler-%d", number)
}

// problemSource carries the inputs needed to assemble a Problem Exercise.
type problemSource struct {
	baseDir  string
	language string
	title    string
	number   int
}

// newProblemFromSource assembles a finished Project Euler Problem Exercise in
// memory. It writes no input (a Problem's input is optional and defaults to
// none) and declares only a single empty Part One test placeholder.
func newProblemFromSource(s problemSource) *Exercise {
	return &Exercise{
		ID:       makeProblemID(s.number),
		Kind:     KindProblem,
		Title:    s.title,
		Language: s.language,
		Number:   s.number,
		Path:     filepath.Join(s.baseDir, strconv.Itoa(s.number)),
		Data: &Data{
			InputData:     "",
			InputFileName: "",
			TestCases: TestCase{
				One: []*Test{{Input: "", Expected: ""}},
				Two: nil,
			},
			Answers: Answer{One: "", Two: ""},
		},
	}
}

// ProblemAdder makes a Project Euler Problem exist in the workspace. It is the
// Euler-kind counterpart to Adder: it builds a Problem Exercise in memory (no
// network, no Page Fetcher) and lays it out via the same Exercise Scaffold.
type ProblemAdder struct {
	scaffold        *exerciseScaffold
	exerciseBaseDir string
	language        string
	title           string
	number          int
	logger          *slog.Logger
	fetcher         *problemFetcher

	path               string
	report             Report
	titlePlaceholdered bool
}

// WithProblemNumber sets the Project Euler problem number.
func WithProblemNumber(n int) func(*ProblemAdder) {
	return func(p *ProblemAdder) { p.number = n }
}

// WithProblemLanguage sets the implementation language.
func WithProblemLanguage(lang string) func(*ProblemAdder) {
	return func(p *ProblemAdder) {
		if lang != "" {
			p.language = lang
		}
	}
}

// WithProblemFetcher overrides the title fetcher. Its parameter is the
// unexported *problemFetcher, so this is effectively a test-only seam for
// injecting an httptest-backed client.
func WithProblemFetcher(f *problemFetcher) func(*ProblemAdder) {
	return func(p *ProblemAdder) { p.fetcher = f }
}

// NewProblemAdder builds a ProblemAdder from config and options, then validates.
func NewProblemAdder(cfg config.Config, opts ...func(*ProblemAdder)) (*ProblemAdder, error) {
	p := &ProblemAdder{
		exerciseBaseDir: cfg.GetEulerDir(),
		language:        cfg.GetLanguage(),
		logger:          cfg.GetLogger(),
		scaffold: &exerciseScaffold{
			fs:            cfg.GetFs(),
			inputFileName: cfg.GetInputFilename(),
			overwrites:    &Overwrites{},
			logger:        cfg.GetLogger(),
		},
	}

	p.fetcher = newProblemFetcher()

	for _, opt := range opts {
		opt(p)
	}

	if p.language == "" {
		return nil, ErrEmptyLanguage
	}
	if p.number <= 0 {
		return nil, fmt.Errorf("problem number must be positive: %w", ErrInvalidData)
	}

	return p, nil
}

// Add resolves the Problem's title from projecteuler.net, then assembles the
// Problem in memory and lays it out on disk.
func (p *ProblemAdder) Add() error {
	title, err := p.fetcher.fetchTitle(p.number)
	switch {
	case err == nil:
		p.title = title
	case errors.Is(err, ErrInvalidData):
		// A fetched page with no title means the number does not exist: hard-fail
		// so a typo'd number leaves nothing scaffolded behind.
		p.logger.Error("problem not found", slog.Int("number", p.number), tint.Err(err))
		return fmt.Errorf("problem %d not found: %w", p.number, err)
	default:
		// The site was unreachable: degrade to a placeholder so an offline add
		// still produces a solvable exercise. The command surfaces the warning.
		p.logger.Warn("title fetch failed; using placeholder",
			slog.Int("number", p.number), slog.String("placeholder", placeholderTitle), tint.Err(err))
		p.title = placeholderTitle
		p.titlePlaceholdered = true
	}

	ex := newProblemFromSource(problemSource{
		baseDir:  p.exerciseBaseDir,
		language: p.language,
		title:    p.title,
		number:   p.number,
	})

	p.path = ex.Path

	report, err := p.scaffold.write(ex)
	if err != nil {
		p.logger.Error("scaffolding problem", slog.Int("number", p.number), tint.Err(err))
		return fmt.Errorf("adding problem: %w", err)
	}
	p.report = report

	return nil
}

// FilePath returns the problem directory path after Add.
func (p *ProblemAdder) FilePath() string { return p.path }

// Report returns the scaffold report after Add.
func (p *ProblemAdder) Report() Report { return p.report }

// TitlePlaceholdered reports whether the title fell back to placeholderTitle
// because projecteuler.net was unreachable. The command uses it to warn the user.
func (p *ProblemAdder) TitlePlaceholdered() bool { return p.titlePlaceholdered }
