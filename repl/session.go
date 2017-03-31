package repl

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type Session struct {
	Imports         []string
	Lines           []string
	originalImports []string
	History         string
	SkipHistory     bool
	Debug           bool
}

func NewSession(imports ...string) *Session {
	s := &Session{
		Imports:         imports,
		Lines:           []string{},
		originalImports: imports,
	}
	return s
}

func (s *Session) AddImports(imports ...string) {
	s.Imports = append(s.Imports, imports...)
}

func (s *Session) AddLines(lines ...string) {
	s.Lines = append(s.Lines, lines...)
}

func (s *Session) Clear() {
	s.Imports = []string{}
	s.Lines = []string{}
}

func (s *Session) Reset() {
	s.Clear()
	s.Imports = s.originalImports
}

func (s *Session) Execute(data string, out io.Writer) ([]byte, error) {
	s.Clear()
	for _, line := range strings.Split(data, "\n") {
		sl := strings.TrimSpace(line)
		if strings.HasPrefix(sl, "//") {
			continue
		}
		if strings.HasPrefix(sl, "import") {
			s.AddImports(line)
			continue
		}
		s.AddLines(line)
	}
	dir, err := ioutil.TempDir(os.TempDir(), "replo")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer os.RemoveAll(dir)

	f, err := os.Create(filepath.Join(dir, "replo.go"))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()

	err = s.Print(f)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	s.goImports(f.Name(), out)
	cmd := exec.Command("go", "run", f.Name())
	return cmd.CombinedOutput()
}

func (s *Session) goImports(path string, out io.Writer) {
	if _, err := exec.LookPath("goimports"); err != nil {
		return
	}
	cmd := exec.Command("goimports", "-w", path)
	cmd.Run()
	if s.Debug {
		if f, err := os.Open(path); err == nil {
			io.Copy(out, f)
		}
	}
}

func (s *Session) Print(out io.Writer) error {
	t, err := template.New("replo").Parse(tmpl)
	if err != nil {
		return errors.WithStack(err)
	}
	return t.Execute(out, s)
}

const tmpl = `package main

{{range .Imports -}}
	{{.}}
{{end }}

func main() {
{{range .Lines -}}
	{{.}}
{{end -}}
}
`
