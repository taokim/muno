# MUNO shell integration for {{CMD_NAME}} (fish)
function {{CMD_NAME}}
    set -l target $argv[1]
    test -z "$target" && set target "."
    
    # Special navigation patterns
    switch $target
        case '-'  # Previous location
            if test -z "$_MUNO_PREV"
                echo "No previous location"
                return 1
            end
            set target $_MUNO_PREV
        case '...'  # Grandparent
            set target "../.."
    end
    
    # Try to find muno binary (prefer muno-local from PATH, fallback to system)
    set -l muno_cmd "muno"
    if command -v muno-local >/dev/null 2>&1
        set muno_cmd "muno-local"
    end
    
    # Save current position
    set -g _MUNO_PREV ($muno_cmd path . --relative 2>/dev/null; or echo '/')
    
    # Resolve path with --ensure to auto-clone lazy repositories
    set -l resolved ($muno_cmd path $target --ensure 2>/dev/null)

    if test $status -eq 0; and test -d "$resolved"
        cd $resolved
        echo "📍 "($muno_cmd path . --relative 2>/dev/null; or pwd)
    else
        echo "❌ Failed to resolve: $target" >&2
        return 1
    end
end

# Completion for {{CMD_NAME}}
function __{{CMD_NAME}}_complete
    # Try to find muno binary (prefer muno-local from PATH, fallback to system)
    set -l muno_cmd "muno"
    if command -v muno-local >/dev/null 2>&1
        set muno_cmd "muno-local"
    end
    
    set -l nodes
    
    # Get node names using the new quiet mode (if available)
    if $muno_cmd list --help 2>/dev/null | grep -q "\-\-quiet"
        set nodes ($muno_cmd list --quiet 2>/dev/null)
    else
        # Fallback: parse regular list output
        set nodes ($muno_cmd list 2>/dev/null | grep -E "^\s*[✅💤]" | sed 's/^[[:space:]]*[✅💤][[:space:]]*//' | sed 's/[[:space:]].*//' | sort -u)
    end
    
    # Add common navigation patterns
    set nodes $nodes . .. /
    
    # Check if current word starts with /
    set -l current_token (commandline -ct)
    if string match -q '/*' $current_token
        # For absolute paths, get all possible tree paths with recursive mode
        if $muno_cmd list --help 2>/dev/null | grep -q "\-\-quiet"
            set -l tree_paths ($muno_cmd list --quiet --recursive 2>/dev/null | sed 's|^|/|')
        else
            # Fallback: parse recursive list output
            set -l tree_paths ($muno_cmd list --recursive 2>/dev/null | grep -E "^\s*[✅💤]" | sed 's/^[[:space:]]*[✅💤][[:space:]]*//' | sed 's/[[:space:]].*//' | sed 's|^|/|' | sort -u)
        end
        set nodes $nodes $tree_paths /
    end
    
    printf '%s\n' $nodes
end

complete -c {{CMD_NAME}} -a '(__{{CMD_NAME}}_complete)'

# Optional aliases
alias {{CMD_NAME}}t='muno tree'
alias {{CMD_NAME}}s='muno status --recursive'
alias {{CMD_NAME}}l='muno list'