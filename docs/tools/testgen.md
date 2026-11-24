# Test Generator

Generate unit tests for Go, TypeScript/JavaScript, and Python files.

## Usage

```bash
# Generate tests for a file (auto-detects framework)
ai-dev testgen ./handler.go

# Generate for multiple files
ai-dev testgen ./pkg/... ./cmd/...

# Specify framework explicitly
ai-dev testgen ./component.tsx --framework jest

# Output to stdout instead of file
ai-dev testgen ./handler.go --output stdout

# Focus on coverage
ai-dev testgen ./handler.go --coverage
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--framework` | `auto` | `gotest`, `jest`, `vitest`, `pytest`, `auto` |
| `--output` | `file` | `file`, `stdout` |
| `--coverage` | `false` | Generate tests with coverage focus |

## Framework Detection

| Files Present | Detected Framework |
|---------------|-------------------|
| `*_test.go` | gotest |
| `jest.config.js`, `jest.config.ts` | jest |
| `vitest.config.ts`, `vitest.config.js` | vitest |
| `*_test.py`, `pytest.ini` | pytest |

## Supported Frameworks

### Go (gotest)
```go
func TestHandler(t *testing.T) {
    testCases := []struct {
        name     string
        input    string
        expected string
    }{
        {"empty", "", "default"},
        {"normal", "hello", "hello world"},
    }
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := Handle(tc.input)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

### TypeScript/Jest
```typescript
describe('Component', () => {
  it('renders correctly', () => {
    render(<Component />);
    expect(screen.getByText('Hello')).toBeInTheDocument();
  });
});
```

### TypeScript/Vitest
```typescript
import { describe, it, expect } from 'vitest';

describe('Component', () => {
  it('renders correctly', () => {
    // ...
  });
});
```

### Python/pytest
```python
import pytest

class TestHandler:
    @pytest.mark.parametrize("input,expected", [
        ("", "default"),
        ("hello", "hello world"),
    ])
    def test_handle(self, input, expected):
        assert handle(input) == expected
```

## Configuration

```yaml
tools:
  testgen:
    framework: auto     # gotest, jest, vitest, pytest, auto
    output: file        # file, stdout
    coverage: false
```

## How It Works

1. Reads source file
2. Detects framework (or uses explicit flag)
3. Finds existing tests to avoid duplication
4. Sends source + context to LLM with test generation prompt
5. Writes generated test file (or prints to stdout)