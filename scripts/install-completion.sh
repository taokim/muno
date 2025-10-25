#!/usr/bin/env bash
# Install MUNO shell completions for bash, zsh, and fish

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
COMPLETIONS_DIR="$SCRIPT_DIR/../completions"

# Detect shell
detect_shell() {
    if [ -n "$BASH_VERSION" ]; then
        echo "bash"
    elif [ -n "$ZSH_VERSION" ]; then
        echo "zsh"
    elif [ -n "$FISH_VERSION" ]; then
        echo "fish"
    else
        # Try to detect from SHELL variable
        case "$SHELL" in
            *bash) echo "bash" ;;
            *zsh) echo "zsh" ;;
            *fish) echo "fish" ;;
            *) echo "unknown" ;;
        esac
    fi
}

install_bash() {
    echo "Installing bash completion..."
    
    # Find completion directory
    if [[ -d "/usr/local/etc/bash_completion.d" ]]; then
        COMP_DIR="/usr/local/etc/bash_completion.d"
    elif [[ -d "$HOME/.local/share/bash-completion/completions" ]]; then
        COMP_DIR="$HOME/.local/share/bash-completion/completions"
    else
        COMP_DIR="$HOME/.bash_completion.d"
        mkdir -p "$COMP_DIR"
    fi
    
    cp "$COMPLETIONS_DIR/muno.bash" "$COMP_DIR/muno"
    echo "✅ Installed to $COMP_DIR/muno"
    
    # Add to bashrc if needed
    RC_FILE="$HOME/.bashrc"
    [[ -f "$HOME/.bash_profile" ]] && RC_FILE="$HOME/.bash_profile"
    
    if ! grep -q "$COMP_DIR/muno" "$RC_FILE" 2>/dev/null; then
        echo "" >> "$RC_FILE"
        echo "# MUNO bash completion" >> "$RC_FILE"
        echo "[[ -f $COMP_DIR/muno ]] && source $COMP_DIR/muno" >> "$RC_FILE"
        echo "✅ Added to $RC_FILE"
    fi
    
    # Add mcd function if not exists
    if ! grep -q "^mcd()" "$RC_FILE" 2>/dev/null; then
        echo "" >> "$RC_FILE"
        echo "# MUNO mcd function" >> "$RC_FILE"
        echo 'mcd() { cd "$(muno path "$@")" && pwd; }' >> "$RC_FILE"
        echo "✅ Added mcd function to $RC_FILE"
    fi
}

install_zsh() {
    echo "Installing zsh completion..."
    
    # Find completion directory
    if [[ -d "/usr/local/share/zsh/site-functions" ]]; then
        COMP_DIR="/usr/local/share/zsh/site-functions"
    elif [[ -d "$HOME/.local/share/zsh/completions" ]]; then
        COMP_DIR="$HOME/.local/share/zsh/completions"
    else
        COMP_DIR="$HOME/.zsh/completions"
        mkdir -p "$COMP_DIR"
    fi
    
    cp "$COMPLETIONS_DIR/muno.zsh" "$COMP_DIR/_muno"
    echo "✅ Installed to $COMP_DIR/_muno"
    
    # Add to zshrc if needed
    RC_FILE="$HOME/.zshrc"
    
    if ! grep -q "fpath.*$COMP_DIR" "$RC_FILE" 2>/dev/null; then
        echo "" >> "$RC_FILE"
        echo "# MUNO zsh completion" >> "$RC_FILE"
        echo "fpath=($COMP_DIR \$fpath)" >> "$RC_FILE"
        echo "autoload -Uz compinit && compinit" >> "$RC_FILE"
        echo "✅ Added to $RC_FILE"
    fi
    
    # Add mcd function if not exists
    if ! grep -q "^mcd()" "$RC_FILE" 2>/dev/null; then
        echo "" >> "$RC_FILE"
        echo "# MUNO mcd function" >> "$RC_FILE"
        echo 'mcd() { cd "$(muno path "$@")" && pwd; }' >> "$RC_FILE"
        echo "✅ Added mcd function to $RC_FILE"
    fi
}

install_fish() {
    echo "Installing fish completion..."
    
    # Find completion directory
    if [[ -d "$HOME/.config/fish/completions" ]]; then
        COMP_DIR="$HOME/.config/fish/completions"
    elif [[ -d "/usr/local/share/fish/vendor_completions.d" ]]; then
        COMP_DIR="/usr/local/share/fish/vendor_completions.d"
    else
        COMP_DIR="$HOME/.config/fish/completions"
        mkdir -p "$COMP_DIR"
    fi
    
    cp "$COMPLETIONS_DIR/muno.fish" "$COMP_DIR/muno.fish"
    echo "✅ Installed to $COMP_DIR/muno.fish"
    
    # Add mcd function if not exists
    CONFIG_FILE="$HOME/.config/fish/config.fish"
    if ! grep -q "^function mcd" "$CONFIG_FILE" 2>/dev/null; then
        echo "" >> "$CONFIG_FILE"
        echo "# MUNO mcd function" >> "$CONFIG_FILE"
        echo "function mcd" >> "$CONFIG_FILE"
        echo '    cd (muno path $argv); and pwd' >> "$CONFIG_FILE"
        echo "end" >> "$CONFIG_FILE"
        echo "✅ Added mcd function to $CONFIG_FILE"
    fi
}

# Main installation
echo "🚀 MUNO Shell Completion Installer"
echo "===================================="

SHELL_TYPE=$(detect_shell)
echo "Detected shell: $SHELL_TYPE"
echo ""

# Allow override
if [[ $# -gt 0 ]]; then
    case "$1" in
        bash|zsh|fish)
            SHELL_TYPE="$1"
            echo "Installing for: $SHELL_TYPE"
            ;;
        all)
            echo "Installing for all shells..."
            install_bash
            install_zsh
            install_fish
            echo ""
            echo "✅ Installation complete for all shells!"
            exit 0
            ;;
        *)
            echo "Usage: $0 [bash|zsh|fish|all]"
            exit 1
            ;;
    esac
fi

# Install for detected/specified shell
case "$SHELL_TYPE" in
    bash)
        install_bash
        echo ""
        echo "✅ Bash installation complete!"
        echo "📌 Reload with: source ~/.bashrc"
        ;;
    zsh)
        install_zsh
        echo ""
        echo "✅ Zsh installation complete!"
        echo "📌 Reload with: source ~/.zshrc"
        ;;
    fish)
        install_fish
        echo ""
        echo "✅ Fish installation complete!"
        echo "📌 Reload with: source ~/.config/fish/config.fish"
        ;;
    *)
        echo "❌ Could not detect shell type"
        echo "Please specify: $0 [bash|zsh|fish|all]"
        exit 1
        ;;
esac

echo ""
echo "📌 Test completion with:"
echo "   muno path <TAB>"
echo "   mcd <TAB>"