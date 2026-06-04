class Keepassview < Formula
  desc "Read-only KeePass (.kdbx) database viewer that runs a local web server"
  homepage "https://github.com/yetanotherchris/keepassview"
  version "VERSION"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yetanotherchris/keepassview/releases/download/vVERSION/keepassview-darwin-arm64.tar.gz"
      sha256 "SHA256"
    else
      url "https://github.com/yetanotherchris/keepassview/releases/download/vVERSION/keepassview-darwin-amd64.tar.gz"
      sha256 "SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yetanotherchris/keepassview/releases/download/vVERSION/keepassview-linux-arm64.tar.gz"
      sha256 "SHA256"
    else
      url "https://github.com/yetanotherchris/keepassview/releases/download/vVERSION/keepassview-linux-amd64.tar.gz"
      sha256 "SHA256"
    end
  end

  def install
    bin.install "keepassview-darwin-arm64" => "keepassview" if OS.mac? && Hardware::CPU.arm?
    bin.install "keepassview-darwin-amd64" => "keepassview" if OS.mac? && !Hardware::CPU.arm?
    bin.install "keepassview-linux-arm64" => "keepassview" if OS.linux? && Hardware::CPU.arm?
    bin.install "keepassview-linux-amd64" => "keepassview" if OS.linux? && !Hardware::CPU.arm?
  end

  test do
    assert_match "keepassview version", shell_output("#{bin}/keepassview --version")
  end
end
