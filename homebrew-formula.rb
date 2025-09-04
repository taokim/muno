# Homebrew Formula for MUNO
# This file should be placed in a tap repository: homebrew-muno/Formula/muno.rb

class Muno < Formula
  desc "Multi-repository orchestration tool with tree-based navigation"
  homepage "https://github.com/taokim/muno"
  url "https://github.com/taokim/muno/archive/refs/tags/v0.9.0.tar.gz"
  sha256 "PLACEHOLDER_SHA256" # Update with actual SHA256 after release
  license "MIT"
  head "https://github.com/taokim/muno.git", branch: "master"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "./cmd/muno"
    
    # Install shell completions
    generate_completions_from_executable(bin/"muno", "completion")
    
    # Install documentation
    doc.install "README.md", "CLAUDE.md"
  end

  test do
    # Test version output
    assert_match version.to_s, shell_output("#{bin}/muno --version 2>&1")
    
    # Test init command
    system "#{bin}/muno", "init", "test-workspace"
    assert_predicate testpath/"muno.yaml", :exist?
    assert_predicate testpath/"nodes", :exist?
  end
end