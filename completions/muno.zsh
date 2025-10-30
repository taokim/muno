#compdef muno
# Zsh completion for MUNO

_muno() {
    local -a commands
    commands=(
        'init:Initialize a new MUNO workspace'
        'add:Add a repository to the workspace'
        'remove:Remove a repository from the workspace'
        'list:List child nodes and repositories'
        'ls:Alias for list'
        'tree:Display repository tree'
        'status:Show git status'
        'pull:Pull changes from repositories'
        'push:Push changes to repositories'
        'path:Resolve tree path to filesystem path'
        'clone:Clone lazy repositories'
        'use:Navigate to a node'
        'commit:Commit changes'
        'agent:Start an AI agent'
        'claude:Start Claude CLI'
        'gemini:Start Gemini CLI'
        'help:Show help'
        'version:Show version'
    )
    
    local -a nodes
    # Get node list, but only if we're in a workspace
    if [[ -f "muno.yaml" ]] || muno path . &>/dev/null; then
        nodes=(${(f)"$(muno list 2>/dev/null | grep -E '^\s+[a-zA-Z0-9]' | awk '{print $1}')"})
    fi
    
    _arguments -C \
        '1: :->command' \
        '*:: :->args'
    
    case $state in
        command)
            _describe 'command' commands
            ;;
        args)
            case $words[1] in
                path)
                    if [[ ${#words[@]} -eq 2 ]]; then
                        # First argument after 'path' - show targets
                        local -a all_targets
                        all_targets=($nodes '.' '..' '/')
                        _describe 'target' all_targets
                    else
                        # Additional arguments - show flags
                        _arguments \
                            '--ensure[Clone lazy repositories if needed]' \
                            '--relative[Show position in tree]' \
                            '--help[Show help]'
                    fi
                    ;;
                add)
                    _arguments \
                        '--name[Custom name for the repository]:name' \
                        '--lazy[Mark as lazy repository]' \
                        '--file[Config file path]:file:_files' \
                        '--help[Show help]' \
                        '1:repository URL'
                    ;;
                remove|use|agent|claude|gemini)
                    _describe 'node' nodes
                    ;;
                pull|push|status|clone)
                    _arguments \
                        '--recursive[Operate recursively]' \
                        '--help[Show help]'
                    ;;
                commit)
                    _arguments \
                        '-m[Commit message]:message' \
                        '--message[Commit message]:message' \
                        '--help[Show help]'
                    ;;
            esac
            ;;
    esac
}

# Completion for mcd function
_mcd() {
    local -a nodes targets
    if [[ -f "muno.yaml" ]] || muno path . &>/dev/null; then
        nodes=(${(f)"$(muno list 2>/dev/null | grep -E '^\s+[a-zA-Z0-9]' | awk '{print $1}')"})
    fi
    targets=($nodes '.' '..' '/' '-' '...')
    # Simple completion without _arguments to avoid conflicts
    compadd -a targets
}

compdef _muno muno
compdef _mcd mcd 2>/dev/null