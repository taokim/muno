#!/usr/bin/env fish
# Fish completion for MUNO

# Helper function to get nodes
function __fish_muno_nodes
    muno list 2>/dev/null | grep -E '^\s+[a-zA-Z0-9]' | awk '{print $1}'
end

# Check if no subcommand has been given
function __fish_muno_needs_command
    set -l cmd (commandline -opc)
    test (count $cmd) -eq 1
end

# Check if specific command was given
function __fish_muno_using_command
    set -l cmd (commandline -opc)
    test (count $cmd) -gt 1
    and test $cmd[2] = $argv[1]
end

# Main commands
complete -f -c muno -n __fish_muno_needs_command -a init -d 'Initialize a new MUNO workspace'
complete -f -c muno -n __fish_muno_needs_command -a add -d 'Add a repository'
complete -f -c muno -n __fish_muno_needs_command -a remove -d 'Remove a repository'
complete -f -c muno -n __fish_muno_needs_command -a list -d 'List child nodes'
complete -f -c muno -n __fish_muno_needs_command -a ls -d 'Alias for list'
complete -f -c muno -n __fish_muno_needs_command -a tree -d 'Display repository tree'
complete -f -c muno -n __fish_muno_needs_command -a status -d 'Show git status'
complete -f -c muno -n __fish_muno_needs_command -a pull -d 'Pull changes'
complete -f -c muno -n __fish_muno_needs_command -a push -d 'Push changes'
complete -f -c muno -n __fish_muno_needs_command -a path -d 'Resolve tree path'
complete -f -c muno -n __fish_muno_needs_command -a clone -d 'Clone lazy repositories'
complete -f -c muno -n __fish_muno_needs_command -a use -d 'Navigate to a node'
complete -f -c muno -n __fish_muno_needs_command -a commit -d 'Commit changes'
complete -f -c muno -n __fish_muno_needs_command -a agent -d 'Start AI agent'
complete -f -c muno -n __fish_muno_needs_command -a claude -d 'Start Claude CLI'
complete -f -c muno -n __fish_muno_needs_command -a gemini -d 'Start Gemini CLI'
complete -f -c muno -n __fish_muno_needs_command -a help -d 'Show help'
complete -f -c muno -n __fish_muno_needs_command -a version -d 'Show version'

# Command-specific completions
# path command
complete -f -c muno -n '__fish_muno_using_command path' -l ensure -d 'Clone lazy repositories if needed'
complete -f -c muno -n '__fish_muno_using_command path' -l relative -d 'Show position in tree'
complete -f -c muno -n '__fish_muno_using_command path' -l help -d 'Show help'
complete -f -c muno -n '__fish_muno_using_command path' -a '(__fish_muno_nodes)' -d 'Node'
complete -f -c muno -n '__fish_muno_using_command path' -a '.' -d 'Current directory'
complete -f -c muno -n '__fish_muno_using_command path' -a '..' -d 'Parent directory'
complete -f -c muno -n '__fish_muno_using_command path' -a '/' -d 'Workspace root'

# add command
complete -f -c muno -n '__fish_muno_using_command add' -l name -d 'Custom name for repository'
complete -f -c muno -n '__fish_muno_using_command add' -l lazy -d 'Mark as lazy repository'
complete -f -c muno -n '__fish_muno_using_command add' -l file -d 'Config file path'
complete -f -c muno -n '__fish_muno_using_command add' -l help -d 'Show help'

# remove/use/agent/claude/gemini commands - complete with nodes
for cmd in remove use agent claude gemini
    complete -f -c muno -n "__fish_muno_using_command $cmd" -a '(__fish_muno_nodes)' -d 'Node'
end

# pull/push/status/clone commands
for cmd in pull push status clone
    complete -f -c muno -n "__fish_muno_using_command $cmd" -l recursive -d 'Operate recursively'
    complete -f -c muno -n "__fish_muno_using_command $cmd" -l help -d 'Show help'
end

# commit command
complete -f -c muno -n '__fish_muno_using_command commit' -s m -l message -d 'Commit message'
complete -f -c muno -n '__fish_muno_using_command commit' -l help -d 'Show help'

# mcd function completion
function __fish_complete_mcd
    __fish_muno_nodes
    echo "."
    echo ".."
    echo "/"
    echo "-"
    echo "..."
end

complete -f -c mcd -a '(__fish_complete_mcd)' -d 'MUNO navigation target'