package workspace

import (
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/alfarax9/astra/internal/domain"
)

// mapTree is a repository given as path -> content.
type mapTree map[string]string

func (m mapTree) List(dir string) ([]string, error) {
	var names []string
	for p := range m {
		rel, ok := strings.CutPrefix(p, dir+"/")
		if dir == "" {
			rel, ok = p, true
		}
		if !ok {
			continue
		}
		name, _, _ := strings.Cut(rel, "/")
		if !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("%s: %w", dir, fs.ErrNotExist)
	}
	slices.Sort(names)
	return names, nil
}

func (m mapTree) Read(p string) ([]byte, error) {
	if c, ok := m[path.Clean(p)]; ok {
		return []byte(c), nil
	}
	return nil, fs.ErrNotExist
}

func pkg(deps, scripts string) string {
	return `{"dependencies": {` + deps + `}, "scripts": {` + scripts + `}}`
}

const nodeScripts = `"dev": "x", "build": "x", "test": "vitest"`

func TestDetectsTheTenSampleApps(t *testing.T) {
	cases := []struct {
		name  string
		repo  mapTree
		stack string
		want  domain.Commands
	}{
		{"Next.js", mapTree{"package.json": pkg(`"next": "15", "react": "19"`, nodeScripts), "package-lock.json": ""},
			"nextjs", domain.Commands{Dev: "npm run dev", Build: "npm run build", Test: "npm run test"}},
		{"Astro", mapTree{"package.json": pkg(`"astro": "5"`, `"dev": "astro dev", "build": "astro build"`), "pnpm-lock.yaml": ""},
			"astro", domain.Commands{Dev: "pnpm run dev", Build: "pnpm run build"}},
		{"Vite", mapTree{"package.json": `{"devDependencies": {"vite": "7"}, "scripts": {` + nodeScripts + `}}`, "yarn.lock": ""},
			"vite", domain.Commands{Dev: "yarn run dev", Build: "yarn run build", Test: "yarn run test"}},
		{"Laravel", mapTree{"artisan": "", "composer.json": `{"require-dev": {"phpunit/phpunit": "11"}}`,
			"package.json": pkg(`"vite": "7"`, `"build": "vite build"`)},
			"laravel", domain.Commands{Dev: "php artisan serve", Build: "npm run build", Test: "php artisan test"}},
		{"Express", mapTree{"package.json": `{"main": "server.js", "dependencies": {"express": "5"}, "scripts": {"test": "echo \"Error: no test specified\" && exit 1"}}`},
			"express", domain.Commands{Dev: "node server.js"}},
		{"FastAPI", mapTree{"requirements.txt": "FastAPI==0.115\nuvicorn\npytest\n", "app/main.py": ""},
			"fastapi", domain.Commands{Dev: "uvicorn app.main:app --reload", Test: "pytest"}},
		{"Go", mapTree{"go.mod": "module x", "cmd/server/main.go": ""},
			"go", domain.Commands{Dev: "go run ./cmd/server", Build: "go build ./...", Test: "go test ./..."}},
		{"Django", mapTree{"manage.py": "", "requirements.txt": "django"},
			"django", domain.Commands{Dev: "python manage.py runserver", Test: "python manage.py test"}},
		{"static HTML", mapTree{"index.html": "<h1>hi</h1>", "style.css": ""}, "static", domain.Commands{}},
		{"custom Dockerfile", mapTree{"Dockerfile": "FROM rust", "Cargo.toml": "", "src/main.rs": ""},
			"dockerfile", domain.Commands{Build: "docker build ."}},
	}
	right := 0
	for _, c := range cases {
		stack, got, err := detect(c.repo)
		if err == nil && stack == c.stack && got == c.want {
			right++
			continue
		}
		t.Errorf("%s: got %q %+v, %v; want %q %+v", c.name, stack, got, err, c.stack, c.want)
	}
	// The PRD asks for at least 9 of 10 (F2-03); every one should be right.
	if right < len(cases) {
		t.Errorf("%d of %d sample apps detected", right, len(cases))
	}
}

func TestDetectsOtherShapes(t *testing.T) {
	cases := []struct {
		name  string
		repo  mapTree
		stack string
		want  domain.Commands
	}{
		{"node with start only", mapTree{"package.json": pkg(``, `"start": "node ."`), "bun.lockb": ""}, "node", domain.Commands{Dev: "bun run start"}},
		{"node with index.js", mapTree{"package.json": `{}`, "index.js": ""}, "node", domain.Commands{Dev: "node index.js"}},
		{"broken package.json", mapTree{"package.json": `{`, "index.html": ""}, "static", domain.Commands{}},
		{"go with main.go", mapTree{"go.mod": "", "main.go": ""}, "go", domain.Commands{Dev: "go run .", Build: "go build ./...", Test: "go test ./..."}},
		{"flask", mapTree{"pyproject.toml": "dependencies = [\"flask\"]", "tests/test_app.py": ""}, "flask", domain.Commands{Dev: "flask run", Test: "pytest"}},
		{"fastapi in main.py", mapTree{"Pipfile": "fastapi", "main.py": ""}, "fastapi", domain.Commands{Dev: "uvicorn main:app --reload"}},
		{"fastapi in app.py", mapTree{"requirements.txt": "fastapi", "app.py": ""}, "fastapi", domain.Commands{Dev: "uvicorn app:app --reload"}},
		{"plain python", mapTree{"app.py": ""}, "python", domain.Commands{Dev: "python app.py"}},
		{"python script", mapTree{"requirements.txt": "requests", "main.py": ""}, "python", domain.Commands{Dev: "python main.py"}},
		{"php with public", mapTree{"composer.json": "{}", "public/index.php": ""}, "php", domain.Commands{Dev: "php -S 0.0.0.0:8000 -t public"}},
		{"php file", mapTree{"index.php": ""}, "php", domain.Commands{Dev: "php -S 0.0.0.0:8000"}},
		{"unknown", mapTree{"README.md": ""}, "", domain.Commands{}},
	}
	for _, c := range cases {
		stack, got, err := detect(c.repo)
		if err != nil || stack != c.stack || got != c.want {
			t.Errorf("%s: got %q %+v, %v; want %q %+v", c.name, stack, got, err, c.stack, c.want)
		}
	}
	if _, _, err := detect(mapTree{}); err == nil {
		t.Error("an unreadable repository must be an error")
	}
}
