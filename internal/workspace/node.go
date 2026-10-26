package workspace

import (
	"strings"

	"github.com/alfarax9/astra/internal/domain"
)

// nodeFrameworks maps a dependency to the stack it means, most specific first.
var nodeFrameworks = []struct{ dep, stack string }{
	{"next", "nextjs"}, {"nuxt", "nuxt"}, {"astro", "astro"}, {"@sveltejs/kit", "sveltekit"},
	{"@remix-run/react", "remix"}, {"vite", "vite"}, {"express", "express"}, {"fastify", "fastify"},
}

// npmPlaceholder is the test script npm init writes; it is not a test suite.
const npmPlaceholder = "no test specified"

func detectNode(t tree, top []string) (string, domain.Commands, bool) {
	pkg, ok := readPackage(t, top)
	if !ok {
		return "", domain.Commands{}, false
	}
	stack := "node"
	for _, f := range nodeFrameworks {
		if pkg.Dependencies[f.dep] != "" || pkg.DevDependencies[f.dep] != "" {
			stack = f.stack
			break
		}
	}
	pm := packageManager(top)
	c := domain.Commands{}
	switch {
	case pkg.Scripts["dev"] != "":
		c.Dev = pm + " run dev"
	case pkg.Scripts["start"] != "":
		c.Dev = pm + " run start"
	case pkg.Main != "":
		c.Dev = "node " + pkg.Main
	case has(top, "index.js"):
		c.Dev = "node index.js"
	}
	if pkg.Scripts["build"] != "" {
		c.Build = pm + " run build"
	}
	if test := pkg.Scripts["test"]; test != "" && !strings.Contains(test, npmPlaceholder) {
		c.Test = pm + " run test"
	}
	return stack, c, true
}

// packageManager picks npm, pnpm, yarn or bun from the lock file in the repository.
func packageManager(top []string) string {
	switch {
	case has(top, "pnpm-lock.yaml"):
		return "pnpm"
	case has(top, "yarn.lock"):
		return "yarn"
	case has(top, "bun.lock"), has(top, "bun.lockb"):
		return "bun"
	}
	return "npm"
}
