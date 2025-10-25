#!/usr/bin/env bash
# Bash completion for MUNO

_muno() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"
    local cmd="${COMP_WORDS[1]}"
    
    # First argument - complete commands
    if [[ $COMP_CWORD -eq 1 ]]; then
        local commands="init add remove list ls tree status pull push path clone use commit agent claude gemini help version"
        COMPREPLY=($(compgen -W "$commands" -- "$cur"))
        return
    fi
    
    # Command-specific completion
    case "$cmd" in
        path)
            # Complete with nodes and special paths
            if [[ "$cur" == -* ]]; then
                COMPREPLY=($(compgen -W "--ensure --relative --help" -- "$cur"))
            else
                local nodes=$(muno list 2>/dev/null | grep -E '^\s+[a-zA-Z0-9]' | awk '{print $1}')
                local special=". .. /"
                
                # Handle absolute paths
                if [[ "$cur" == /* ]]; then
                    local abs_nodes=""
                    for node in $nodes; do
                        abs_nodes="$abs_nodes /$node"
                    done
                    COMPREPLY=($(compgen -W "$abs_nodes /" -- "$cur"))
                else
                    COMPREPLY=($(compgen -W "$nodes $special" -- "$cur"))
                fi
            fi
            ;;
        add)
            [[ "$cur" == -* ]] && COMPREPLY=($(compgen -W "--name --lazy --file --help" -- "$cur"))
            ;;
        remove|use|agent|claude|gemini)
            # Complete with node names
            local nodes=$(muno list 2>/dev/null | grep -E '^\s+[a-zA-Z0-9]' | awk '{print $1}')
            COMPREPLY=($(compgen -W "$nodes" -- "$cur"))
            ;;
        pull|push|status|clone)
            [[ "$cur" == -* ]] && COMPREPLY=($(compgen -W "--recursive --help" -- "$cur"))
            ;;
        commit)
            [[ "$cur" == -* ]] && COMPREPLY=($(compgen -W "-m --message --help" -- "$cur"))
            ;;
    esac
}

complete -F _muno muno

# Completion for mcd if it exists
_mcd() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local nodes=$(muno list 2>/dev/null | grep -E '^\s+[a-zA-Z0-9]' | awk '{print $1}')
    local special=". .. / - ..."
    
    if [[ "$cur" == /* ]]; then
        local abs_nodes=""
        for node in $nodes; do
            abs_nodes="$abs_nodes /$node"
        done
        COMPREPLY=($(compgen -W "$abs_nodes /" -- "$cur"))
    else
        COMPREPLY=($(compgen -W "$nodes $special" -- "$cur"))
    fi
}

complete -F _mcd mcd 2>/dev/null