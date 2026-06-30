package runtime

import (
	"path/filepath"
	"strings"
)

type Classification struct {
	Runtime    string
	Service    string
	Confidence string
	Evidence   []string
}

func Classify(executable string, args []string, files []string) Classification {
	tokens := make([]string, 0, 1+len(args)+len(files))
	tokens = append(tokens, strings.ToLower(filepath.Base(executable)))
	for _, arg := range args {
		tokens = append(tokens, strings.ToLower(arg))
	}
	for _, file := range files {
		tokens = append(tokens, strings.ToLower(filepath.Base(file)))
	}
	joined := strings.Join(tokens, " ")

	switch {
	case hasAny(joined, "vite", "next", "nuxt", "npm", "pnpm", "yarn", "bun", "node", "package.json"):
		return classification("node", serviceFromNode(joined), "high", "Node.js development marker")
	case hasAny(joined, "artisan", "symfony", "php", "composer.json"):
		return classification("php", "app", "medium", "PHP development marker")
	case hasAny(joined, "uvicorn", "gunicorn", "flask", "django", "fastapi", "pyproject.toml", "requirements.txt"):
		return classification("python", "api", "medium", "Python development marker")
	case hasAny(joined, "go run", "go.mod"):
		return classification("go", "api", "medium", "Go development marker")
	case hasAny(joined, "spring", "gradle", "maven", "java", "pom.xml", "build.gradle", "build.gradle.kts"):
		return classification("java", "api", "medium", "Java development marker")
	default:
		return classification("unknown", "app", "low", "listening local process")
	}
}

func classification(runtimeName, serviceName, confidence, evidence string) Classification {
	return Classification{
		Runtime:    runtimeName,
		Service:    serviceName,
		Confidence: confidence,
		Evidence:   []string{evidence},
	}
}

func serviceFromNode(joined string) string {
	switch {
	case strings.Contains(joined, "next") || strings.Contains(joined, "nuxt") || strings.Contains(joined, "vite"):
		return "app"
	default:
		return "app"
	}
}

func hasAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
