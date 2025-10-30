# MUNO shell integration for {{CMD_NAME}} (zsh)
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
    
    # Try to find muno binary (prefer muno-local from PATH, fallback to system)
    local muno_cmd="muno"
    if command -v muno-local >/dev/null 2>&1; then
        muno_cmd="muno-local"
    fi
    
    # Resolve path (try without lazy clone first)
    local resolved
    resolved=$($muno_cmd path "$target" 2>/dev/null)
    
    # If path doesn't exist, try with lazy clone
    if [ $? -ne 0 ] || [ ! -d "$resolved" ]; then
        resolved=$($muno_cmd path "$target" --ensure 2>/dev/null)
    fi
    
    if [ $? -eq 0 ] && [ -d "$resolved" ]; then
        cd "$resolved"
        # Show current position in tree
        echo "📍 $($muno_cmd path . --relative 2>/dev/null || pwd)"
    else
        echo "❌ Failed to resolve: $target" >&2
        return 1
    fi
}

# Zsh completion support for {{CMD_NAME}}
_{{CMD_NAME}}() {
    local -a nodes
    local current_word="${words[CURRENT]}"
    
    # Try to find muno binary (prefer muno-local from PATH, fallback to system)
    local muno_cmd="muno"
    if command -v muno-local >/dev/null 2>&1; then
        muno_cmd="muno-local"
    fi
    
    # Get node names using the new quiet mode (if available)
    if $muno_cmd list --help 2>/dev/null | grep -q "\\-\\-quiet"; then
        nodes=($($muno_cmd list --quiet 2>/dev/null))
    else
        # Fallback: parse regular list output
        nodes=($($muno_cmd list 2>/dev/null | grep -E "^\\s*[✅💤]" | sed 's/^[[:space:]]*[✅💤][[:space:]]*//' | sed 's/[[:space:]].*//' | sort -u))
    fi
    
    # Add common navigation patterns
    nodes+=(. .. /)
    
    # If current word starts with /, try to get recursive paths
    if [[ "$current_word" == /* ]]; then
        # For absolute paths, get all possible tree paths with recursive mode
        if $muno_cmd list --help 2>/dev/null | grep -q "\\-\\-quiet"; then
            local tree_paths=($($muno_cmd list --quiet --recursive 2>/dev/null | sed 's|^|/|'))
        else
            # Fallback: parse recursive list output
            local tree_paths=($($muno_cmd list --recursive 2>/dev/null | grep -E "^\\s*[✅💤]" | sed 's/^[[:space:]]*[✅💤][[:space:]]*//' | sed 's/[[:space:]].*//' | sed 's|^|/|' | sort -u))
        fi
        nodes+=(${tree_paths[@]})
    fi
    
    # Use compadd directly to avoid argument conflicts
    compadd -a nodes
}

# Enable completion system if not already enabled
autoload -U compinit
compinit

# Register the completion function
compdef _{{CMD_NAME}} {{CMD_NAME}}

# Optional aliases
alias {{CMD_NAME}}t='muno tree'
alias {{CMD_NAME}}s='muno status --recursive'
alias {{CMD_NAME}}l='muno list'