#!/usr/bin/env bash
# Anti-bloat budgets for Astra (docs/go-rules.md). Called by `make check`.
set -euo pipefail

fail=0
report() { echo "budget: $*"; fail=1; }

# Only our packages: ./... would also pick up Go files inside web/node_modules.
pkgs=()
for d in cmd internal; do if [[ -d $d ]]; then pkgs+=("./$d/..."); fi; done

# go list runs here, not inside `< <(...)`, so a failure (an import cycle, a broken
# package) stops the script instead of silently skipping the checks below.
all_pkgs=""; internal_pkgs=""
if (( ${#pkgs[@]} )); then all_pkgs=$(go list -f '{{.Name}} {{.ImportPath}}' "${pkgs[@]}"); fi
if [[ -d internal ]]; then
  internal_dirs=$(go list -f '{{.Dir}}' ./internal/...)
  internal_pkgs=$(go list -f '{{.ImportPath}} {{join .Imports " "}}' ./internal/...)
fi

# 1. File length: <= 400 lines, excluding tests, testdata and generated code.
while IFS= read -r f; do
  if head -n 5 "$f" | grep -q '^// Code generated .* DO NOT EDIT\.$'; then continue; fi
  lines=$(wc -l < "$f")
  if (( lines > 400 )); then report "$f has $lines lines (max 400)"; fi
done < <(git ls-files --cached --others --exclude-standard -- '*.go' | grep -v '_test\.go$' | grep -v '/testdata/' || true)

# 2. Banned package names.
banned='^(util|utils|common|helpers|misc|shared|base)$'
while IFS= read -r line; do
  [[ -n $line ]] || continue
  name=${line%% *}; path=${line#* }
  if [[ $name =~ $banned ]]; then report "package '$name' is not allowed ($path)"; fi
done <<< "$all_pkgs"

# 3. Every package under internal/ has doc.go.
while IFS= read -r dir; do
  [[ -n $dir ]] || continue
  if [[ ! -f "$dir/doc.go" ]]; then report "$dir has no doc.go"; fi
done <<< "${internal_dirs:-}"

# 4. Direct dependencies of the daemon module: <= 25.
direct=$(go list -m -f '{{if and (not .Main) (not .Indirect)}}{{.Path}}{{end}}' all | grep -c . || true)
if (( direct > 25 )); then report "$direct direct dependencies (max 25)"; fi

# 5. TODO must reference an issue: TODO(#123). --untracked also covers files not yet committed.
if git grep --untracked -nE '//[[:space:]]*TODO' -- '*.go' | grep -vE 'TODO\(#[0-9]+\)'; then
  report "TODO without issue number (use TODO(#123): ...)"
fi

# 6. Coverage per capability and domain package: >= 70%.
caps=(agent workspace preview feedback gate deploy watch server domain)
for p in "${caps[@]}"; do
  if [[ ! -d "internal/$p" ]]; then continue; fi
  pct=$(go test -cover "./internal/$p/..." 2>/dev/null \
        | grep -oE 'coverage: [0-9.]+%' | grep -oE '[0-9.]+' | sort -n | head -n1 || true)
  if [[ -z "$pct" ]]; then report "internal/$p has no tests"; continue; fi
  if awk -v v="$pct" 'BEGIN { exit !(v < 70) }'; then report "internal/$p coverage $pct% (min 70%)"; fi
done

# 7. Daemon binary size: at most 5% above budgets/binary-size.txt.
if [[ -f budgets/binary-size.txt ]]; then
  tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags='-s -w' -o "$tmp/astrad" ./cmd/astrad
  size=$(wc -c < "$tmp/astrad"); base=$(< budgets/binary-size.txt)
  limit=$(( base * 105 / 100 ))
  if (( size > limit )); then report "astrad is $size bytes (limit $limit, baseline $base)"; fi
fi

# 8. Infrastructure packages never import capabilities or each other.
mod=$(go list -m)
infra=(store proxy pty agentcli docker dns remote)
caps_re="^$mod/internal/(agent|workspace|preview|feedback|gate|deploy|watch|server|api)(/|$)"
infra_re="^$mod/internal/($(IFS='|'; echo "${infra[*]}"))(/|$)"
while IFS= read -r line; do
  [[ -n $line ]] || continue
  pkg=${line%% *}; imports=${line#* }
  if [[ ! $pkg =~ $infra_re ]]; then continue; fi
  own=$(sed -E "s#^$mod/internal/([^/]+).*#\1#" <<< "$pkg")
  for imp in $imports; do
    if [[ $imp =~ $caps_re ]]; then
      report "$pkg imports $imp (infrastructure is a leaf)"
    elif [[ $imp =~ $infra_re ]] && [[ $imp != "$mod/internal/$own"* ]]; then
      report "$pkg imports $imp (infrastructure packages do not import each other)"
    fi
  done
done <<< "$internal_pkgs"

exit "$fail"
