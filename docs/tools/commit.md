# Commit Generator

Generate conventional commit messages from staged changes.

## Usage

```bash
# Generate and preview (dry-run)
ai-dev commit --dry-run

# Generate and commit interactively
ai-dev commit

# Generate with gitmoji style
ai-dev commit --style gitmoji

# Add Signed-off-by trailer
ai-dev commit --sign-off
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--dry-run` | `-d` | `false` | Print message without committing |
| `--style` | `-s` | `conventional` | `conventional` or `gitmoji` |
| `--sign-off` | | `false` | Add Signed-off-by trailer |

## Conventional Commits Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `revert`

## Examples

```bash
# Preview message
$ ai-dev commit --dry-run
feat(cli): add config init command

# Commit with sign-off
$ ai-dev commit --sign-off
```

## Configuration

```yaml
tools:
  commit:
    style: conventional  # conventional, gitmoji
```

## How It Works

1. Gets staged diff from git
2. Sends diff to LLM with commit prompt template
3. Parses structured JSON response
4. Formats as conventional commit message
5. Creates commit (or prints with `--dry-run`)