# Shell Integration Templates

This directory contains embedded shell script templates for the `muno shell-init` command.

## Structure

```
resources/
└── shell-init/
    ├── bash.sh   # Bash shell integration template
    ├── zsh.sh    # Zsh shell integration template
    └── fish.sh   # Fish shell integration template
```

## Template Variables

All templates use the following placeholder variable:
- `{{CMD_NAME}}` - The command name (default: `mcd`)

## Usage

These templates are embedded into the binary at compile time using Go's `embed` directive.
The templates are processed by:
1. `resources.go` - Embeds the templates using `//go:embed`
2. `getShellTemplate()` - Returns the appropriate template for the shell type
3. `renderShellTemplate()` - Replaces template variables with actual values

## Features

Each template provides:
- Navigation function with special patterns (-, ...)
- Shell-specific tab completion
- Optional aliases for common commands
- Lazy repository cloning support

## Updating Templates

To update a template:
1. Edit the appropriate `.sh` file in `resources/shell-init/`
2. Rebuild the binary with `go build` or `make build`
3. The new template will be embedded automatically

## Testing

After updating templates, test with:
```bash
# Generate scripts for different shells
muno shell-init --shell bash
muno shell-init --shell zsh  
muno shell-init --shell fish

# Test with custom command names
muno shell-init --cmd-name goto

# Test installation/update
muno shell-init --install
muno shell-init --install --force  # For updates
```