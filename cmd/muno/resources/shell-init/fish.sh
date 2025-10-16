# MUNO shell integration for {{CMD_NAME}}
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
    
    # Save current position
    set -g _MUNO_PREV (muno path . --relative 2>/dev/null; or echo '/')
    
    # Resolve path with lazy clone
    set -l resolved (muno path $target --ensure 2>/dev/null)
    
    if test $status -eq 0
        cd $resolved
        echo "📍 "(muno path . --relative 2>/dev/null; or pwd)
    else
        echo "❌ Failed to resolve: $target" >&2
        return 1
    end
end

# Completion for {{CMD_NAME}}
complete -c {{CMD_NAME}} -a '(muno list --format simple 2>/dev/null)'

# Optional aliases
alias {{CMD_NAME}}t='muno tree'
alias {{CMD_NAME}}s='muno status --recursive'
alias {{CMD_NAME}}l='muno list'