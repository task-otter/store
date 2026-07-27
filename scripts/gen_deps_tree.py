#!/usr/bin/env python3
"""Generate deps-tree.md from .deps.yml."""
from __future__ import annotations

import re
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DEPS_FILE = ROOT / ".deps.yml"
OUT_FILE = ROOT / "deps-tree.md"

JS_TOOLS = [
    "biome",
    "bruno",
    "depcheck",
    "eslint",
    "knip",
    "prettier",
    "stylelint",
    "typescript",
]

NODE_ROOTS = {"fnm", "nvm", "bun"}
NODE_PM = {
    "corepack/fnm",
    "corepack/nvm",
    "npm/fnm",
    "npm/nvm",
    "pnpm/fnm",
    "pnpm/nvm",
    "yarn/fnm",
    "yarn/nvm",
}


def parse_deps(path: Path) -> dict[str, list[str]]:
    deps: dict[str, list[str]] = {}
    pattern = re.compile(r"^(\S+):\s*(?:\[(.*)\])?\s*$")
    for line in path.read_text().splitlines():
        if line.strip() in {"", "---"}:
            continue
        match = pattern.match(line)
        if not match:
            raise SystemExit(f"unparseable line in {path}: {line!r}")
        name, raw = match.group(1), match.group(2)
        items = [item.strip() for item in raw.split(",")] if raw else []
        deps[name] = items
    return deps


def depth_of(name: str, deps: dict[str, list[str]], cache: dict[str, int]) -> int:
    if name in cache:
        return cache[name]
    direct = deps.get(name, [])
    if not direct:
        cache[name] = 0
        return 0
    value = 1 + max(depth_of(dep, deps, cache) for dep in direct)
    cache[name] = value
    return value


def module_link(name: str) -> str:
    # internal/* modules are shared helpers under taskfiles/internal/ with no README.
    if name.startswith("internal/"):
        return f"taskfiles/{name}/Taskfile.yml"
    return f"taskfiles/{name}/README.md"


def is_js_variant(name: str) -> bool:
    return any(name.startswith(f"{tool}-") for tool in JS_TOOLS)


def categorize(name: str) -> str:
    if not deps_map.get(name):
        if name in NODE_ROOTS:
            return "node"
        return "standalone"
    if name in NODE_PM or name in NODE_ROOTS or is_js_variant(name):
        return "node"
    if name in {"ansible", "python", "sqlfluff", "yamllint"} or deps_map.get(name) == ["uv"]:
        return "other"
    if name in {"proto", "staticcheck"} or deps_map.get(name) == ["go"]:
        return "other"
    if name in {"gh", "git", "vault"}:
        return "other"
    return "other"


def render_tree(
    name: str,
    deps: dict[str, list[str]],
    prefix: str = "",
    connector: str = "└── ",
    visited: frozenset[str] | None = None,
) -> list[str]:
    visited = visited or frozenset()
    lines = [f"{prefix}{connector}{name}" if prefix else name]
    if name in visited:
        lines.append(f"{prefix}    (cycle)")
        return lines

    direct = deps.get(name, [])
    child_prefix = prefix + ("    " if connector == "└── " else "│   ")
    for index, dep in enumerate(sorted(direct)):
        is_last = index == len(direct) - 1
        branch = "└── " if is_last else "├── "
        lines.extend(
            render_tree(
                dep,
                deps,
                child_prefix,
                branch,
                visited | {name},
            )
        )
    return lines


def build_reverse(deps: dict[str, list[str]]) -> dict[str, list[str]]:
    reverse: dict[str, list[str]] = defaultdict(list)
    for module, module_deps in deps.items():
        for dep in module_deps:
            reverse[dep].append(module)
    for values in reverse.values():
        values.sort()
    return dict(reverse)


def section_forward_by_depth(modules: list[str], deps: dict[str, list[str]]) -> list[str]:
    cache: dict[str, int] = {}
    by_depth: dict[int, list[str]] = defaultdict(list)
    for name in modules:
        by_depth[depth_of(name, deps, cache)].append(name)

    lines: list[str] = []
    for depth in sorted(by_depth):
        names = sorted(by_depth[depth])
        lines.append(f"### Depth {depth}")
        lines.append("")
        for name in names:
            direct = deps.get(name, [])
            if direct:
                lines.append(f"- `{name}` → {', '.join(f'`{dep}`' for dep in direct)}")
            else:
                lines.append(f"- `{name}`")
        lines.append("")
    return lines


def section_forward_trees(modules: list[str], deps: dict[str, list[str]]) -> list[str]:
    lines: list[str] = []
    js_by_pm: dict[str, list[str]] = defaultdict(list)
    other_modules: list[str] = []

    for name in sorted(modules):
        if is_js_variant(name):
            for suffix in ("npm-fnm", "npm-nvm", "pnpm-fnm", "pnpm-nvm", "yarn-fnm", "yarn-nvm", "bun"):
                if name.endswith(f"-{suffix}"):
                    js_by_pm[suffix].append(name)
                    break
        else:
            other_modules.append(name)

    if js_by_pm:
        lines.append("### JS tool stacks")
        lines.append("")
        for suffix in ("npm-fnm", "npm-nvm", "pnpm-fnm", "pnpm-nvm", "yarn-fnm", "yarn-nvm", "bun"):
            variants = js_by_pm.get(suffix, [])
            if not variants:
                continue
            example = variants[0]
            pm = suffix if suffix != "bun" else "bun"
            lines.append(f"**`{pm}` stack** — {len(variants)} modules (`{example}`, …)")
            lines.append("")
            lines.append("```")
            lines.extend(render_tree(example, deps))
            lines.append("```")
            lines.append("")
            lines.append(", ".join(f"`{name}`" for name in variants))
            lines.append("")

    for name in other_modules:
        lines.append(f"**`{name}`**")
        lines.append("")
        lines.append("```")
        lines.extend(render_tree(name, deps))
        lines.append("```")
        lines.append("")

    return lines


def section_reverse(deps: dict[str, list[str]]) -> list[str]:
    reverse = build_reverse(deps)
    lines: list[str] = ["## Reverse tree", ""]
    lines.append("For each module, modules that depend on it (direct dependents only).")
    lines.append("")

    for name in sorted(deps):
        dependents = reverse.get(name, [])
        if not dependents:
            lines.append(f"- `{name}` — *(none)*")
        else:
            joined = ", ".join(f"`{dep}`" for dep in dependents)
            lines.append(f"- `{name}` ← {joined}")
    lines.append("")
    return lines


def generate(deps: dict[str, list[str]]) -> str:
    global deps_map
    deps_map = deps

    standalone = sorted(name for name, module_deps in deps.items() if not module_deps)
    node_modules = sorted(
        name
        for name in deps
        if categorize(name) == "node" and (deps[name] or name in NODE_ROOTS)
    )
    other_modules = sorted(
        name for name in deps if categorize(name) == "other" and deps[name]
    )

    lines = [
        "# Module dependency tree",
        "",
        "Auto-generated from [`.deps.yml`](.deps.yml).",
        "",
        "Regenerate:",
        "",
        "```sh",
        "python3 scripts/gen_deps_tree.py",
        "```",
        "",
        f"**{len(deps)} modules** total.",
        "",
        "## Standalone",
        "",
        "Modules with no `includes:` dependencies.",
        "",
    ]
    lines.extend(f"- [`{name}`]({module_link(name)})" for name in standalone)
    lines.extend(["", "## Forward tree", ""])

    lines.append("### Node.js stacks")
    lines.append("")
    lines.extend(section_forward_by_depth(node_modules, deps))
    lines.extend(section_forward_trees(node_modules, deps))

    lines.append("### Other chains")
    lines.append("")
    lines.extend(section_forward_by_depth(other_modules, deps))
    lines.extend(section_forward_trees(other_modules, deps))

    lines.extend(section_reverse(deps))
    return "\n".join(lines).rstrip() + "\n"


deps_map: dict[str, list[str]] = {}


def main() -> None:
    global deps_map
    deps_map = parse_deps(DEPS_FILE)
    OUT_FILE.write_text(generate(deps_map))
    print(f"Wrote {OUT_FILE.relative_to(ROOT)} ({len(deps_map)} modules)")


if __name__ == "__main__":
    main()
