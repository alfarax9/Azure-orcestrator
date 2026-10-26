package workspace

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/alfarax9/astra/internal/domain"
)

// tree is a read-only view of a repository: List names the entries of a folder ("" for the
// top), Read returns a file.
type tree interface {
	List(dir string) ([]string, error)
	Read(path string) ([]byte, error)
}

// detector recognises one kind of app from the top-level names of a repository.
type detector func(t tree, top []string) (stack string, c domain.Commands, ok bool)

// detectors run in order; the first that recognises the repository wins. Frameworks with
// their own runner (Laravel's artisan, Django's manage.py) come before the language they
// are written in, and a bare Dockerfile or index.html comes last.
var detectors = []detector{detectLaravel, detectDjango, detectNode, detectGo, detectPython, detectPHP, detectDockerfile, detectStatic}

// detect finds the kind of app and the commands to develop, build and test it (F2-03).
// An unknown repository gives an empty stack and no commands; the user fills them in.
func detect(t tree) (string, domain.Commands, error) {
	top, err := t.List("")
	if err != nil {
		return "", domain.Commands{}, err
	}
	for _, d := range detectors {
		if stack, c, ok := d(t, top); ok {
			return stack, c, nil
		}
	}
	return "", domain.Commands{}, nil
}

func has(names []string, name string) bool { return slices.Contains(names, name) }

// readLower returns a file in lower case, or "" when it cannot be read.
func readLower(t tree, path string) string {
	b, err := t.Read(path)
	if err != nil {
		return ""
	}
	return strings.ToLower(string(b))
}

func detectLaravel(t tree, top []string) (string, domain.Commands, bool) {
	if !has(top, "artisan") || !has(top, "composer.json") {
		return "", domain.Commands{}, false
	}
	c := domain.Commands{Dev: "php artisan serve", Test: "php artisan test"}
	if pkg, ok := readPackage(t, top); ok && pkg.Scripts["build"] != "" {
		c.Build = packageManager(top) + " run build"
	}
	return "laravel", c, true
}

func detectDjango(_ tree, top []string) (string, domain.Commands, bool) {
	if !has(top, "manage.py") {
		return "", domain.Commands{}, false
	}
	return "django", domain.Commands{Dev: "python manage.py runserver", Test: "python manage.py test"}, true
}

func detectGo(t tree, top []string) (string, domain.Commands, bool) {
	if !has(top, "go.mod") {
		return "", domain.Commands{}, false
	}
	c := domain.Commands{Dev: "go run .", Build: "go build ./...", Test: "go test ./..."}
	if !has(top, "main.go") && has(top, "cmd") {
		if cmds, err := t.List("cmd"); err == nil && len(cmds) > 0 {
			c.Dev = "go run ./cmd/" + cmds[0]
		}
	}
	return "go", c, true
}

func detectPython(t tree, top []string) (string, domain.Commands, bool) {
	var deps string
	for _, f := range []string{"pyproject.toml", "requirements.txt", "Pipfile"} {
		if has(top, f) {
			deps += readLower(t, f)
		}
	}
	if deps == "" && !has(top, "main.py") && !has(top, "app.py") {
		return "", domain.Commands{}, false
	}
	c := domain.Commands{}
	if strings.Contains(deps, "pytest") || has(top, "tests") {
		c.Test = "pytest"
	}
	switch {
	case strings.Contains(deps, "fastapi"):
		c.Dev = "uvicorn " + asgiModule(t, top) + ":app --reload"
		return "fastapi", c, true
	case strings.Contains(deps, "flask"):
		c.Dev = "flask run"
		return "flask", c, true
	case has(top, "main.py"):
		c.Dev = "python main.py"
	case has(top, "app.py"):
		c.Dev = "python app.py"
	}
	return "python", c, true
}

// asgiModule finds the module that holds a FastAPI app: main.py, app.py or app/main.py.
func asgiModule(t tree, top []string) string {
	switch {
	case has(top, "main.py"):
		return "main"
	case has(top, "app.py"):
		return "app"
	}
	if inner, err := t.List("app"); err == nil && has(inner, "main.py") {
		return "app.main"
	}
	return "main"
}

func detectPHP(t tree, top []string) (string, domain.Commands, bool) {
	if !has(top, "composer.json") && !has(top, "index.php") {
		return "", domain.Commands{}, false
	}
	c := domain.Commands{Dev: "php -S 0.0.0.0:8000"}
	if has(top, "public") {
		c.Dev += " -t public"
	}
	if strings.Contains(readLower(t, "composer.json"), "phpunit/phpunit") {
		c.Test = "vendor/bin/phpunit"
	}
	return "php", c, true
}

func detectDockerfile(_ tree, top []string) (string, domain.Commands, bool) {
	if !has(top, "Dockerfile") {
		return "", domain.Commands{}, false
	}
	return "dockerfile", domain.Commands{Build: "docker build ."}, true
}

// detectStatic recognises a site of plain files; previews serve them without a command.
func detectStatic(_ tree, top []string) (string, domain.Commands, bool) {
	return "static", domain.Commands{}, has(top, "index.html")
}

type packageJSON struct {
	Main            string            `json:"main"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func readPackage(t tree, top []string) (packageJSON, bool) {
	var pkg packageJSON
	if !has(top, "package.json") {
		return pkg, false
	}
	b, err := t.Read("package.json")
	if err != nil || json.Unmarshal(b, &pkg) != nil {
		return pkg, false
	}
	return pkg, true
}
