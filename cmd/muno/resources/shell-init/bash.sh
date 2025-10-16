# MUNO shell integration for {{CMD_NAME}}
{{CMD_NAME}}() {
    local target="${1:-.}"
    
    # Special navigation patterns
    case "$target" in
        -)  # Previous location
            target="${_MUNO_PREV:-}"
            [ -z "$target" ] && echo "No previous location" && return 1
            ;;
        ...)  # Grandparent  
            target="../.."
            ;;
    esac
    
    # Save current position for '-' navigation
    _MUNO_PREV="$(muno path . --relative 2>/dev/null || echo '/')"
    
    # Resolve path with lazy clone
    local resolved
    resolved=$(muno path "$target" --ensure 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        cd "$resolved"
        # Show current position in tree
        echo "📍 $(muno path . --relative 2>/dev/null || pwd)"
    else
        echo "❌ Failed to resolve: $target" >&2
        return 1
    fi
}

# Completion support for {{CMD_NAME}}
_{{CMD_NAME}}_complete() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local nodes=$(muno list --format simple 2>/dev/null | grep -v "^$")
    COMPREPLY=($(compgen -W "$nodes" -- "$cur"))
}
complete -F _{{CMD_NAME}}_complete {{CMD_NAME}}

# Optional aliases
alias {{CMD_NAME}}t='muno tree'
alias {{CMD_NAME}}s='muno status --recursive'
alias {{CMD_NAME}}l='muno list'