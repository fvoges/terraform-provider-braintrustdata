#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

failures=0

report_failure() {
  echo "[examples-lint] $1"
  failures=1
}

check_forbidden_member_ids() {
  local hits
  hits="$(rg -n "\\bmember_ids\\b" examples docs || true)"
  if [ -n "$hits" ]; then
    report_failure "Found forbidden token 'member_ids'. Use member_users/member_groups instead."
    echo "$hits"
  fi
}

check_leaf_contracts() {
  local dir
  for dir in examples/resources/* examples/data-sources/*; do
    [ -d "$dir" ] || continue

    if [ ! -f "$dir/README.md" ]; then
      report_failure "Missing README.md in $dir"
    fi

    if [ ! -f "$dir/versions.tf" ]; then
      report_failure "Missing versions.tf in $dir"
    fi
  done
}

check_import_targets() {
  local script dir tf_file address type name
  for script in examples/resources/*/import.sh; do
    [ -f "$script" ] || continue
    dir="$(dirname "$script")"
    tf_file="$dir/resource.tf"

    if [ ! -f "$tf_file" ]; then
      report_failure "Missing resource.tf paired with import script: $script"
      continue
    fi

    while read -r address; do
      [ -n "$address" ] || continue
      type="${address%%.*}"
      name="${address#*.}"

      if ! rg -q "resource[[:space:]]+\"${type}\"[[:space:]]+\"${name}\"" "$tf_file"; then
        report_failure "Import target ${address} in ${script} does not exist in ${tf_file}"
      fi
    done < <(awk '$1=="terraform" && $2=="import" { print $3 }' "$script")
  done
}

check_lockfiles_policy() {
  local lockfiles
  lockfiles="$(find examples -type f -name '.terraform.lock.hcl' | sort || true)"
  if [ -n "$lockfiles" ]; then
    report_failure "Lockfiles are not tracked in examples/. Remove these files:"
    echo "$lockfiles"
  fi
}

check_forbidden_member_ids
check_leaf_contracts
check_import_targets
check_lockfiles_policy

if [ "$failures" -ne 0 ]; then
  echo "[examples-lint] FAILED"
  exit 1
fi

echo "[examples-lint] PASSED"
