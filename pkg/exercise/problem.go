package exercise

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"

	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/protocol"
)

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

	path   string
	report Report
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

// WithProblemTitle sets the human-readable problem title.
func WithProblemTitle(title string) func(*ProblemAdder) {
	return func(p *ProblemAdder) { p.title = title }
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

	for _, opt := range opts {
		opt(p)
	}

	if p.language == "" {
		return nil, ErrEmptyLanguage
	}
	if p.number <= 0 {
		return nil, fmt.Errorf("problem number must be positive: %w", ErrInvalidData)
	}
	if p.title == "" {
		return nil, fmt.Errorf("problem title is required: %w", ErrInvalidData)
	}

	return p, nil
}

// Add assembles the Problem in memory and lays it out on disk.
func (p *ProblemAdder) Add() error {
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
