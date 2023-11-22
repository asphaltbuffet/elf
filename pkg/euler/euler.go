package euler

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

type Problem struct {
	ID       int
	Language string
	Runner   runners.Runner
}

func New(id int, lang string) *Problem {
	return &Problem{
		ID:       id,
		Language: lang,
		Runner:   runners.Available[lang](filepath.Join("problems", fmt.Sprintf("%03d", id), lang)),
	}
}

func (p *Problem) SetLanguage(lang string) {
	slog.Debug("setting language", slog.String("language", lang))
	p.Language = lang
}

func (p *Problem) Dir() string {
	return fmt.Sprintf("problems/%03d", p.ID)
}
