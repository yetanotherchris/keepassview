class Keepassview < Formula
  desc "Read-only KeePass (.kdbx) database viewer that runs a local web server"
  homepage "https://github.com/yetanotherchris/keepassview"
  version "1.2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yetanotherchris/keepassview/releases/download/v1.2.0/keepassview-darwin-arm64.tar.gz"
      sha256 "facd09ed3e85093f7798f88e37cccd8ce415847c49703a13f07426960d67ab37"
    else
      url "https://github.com/yetanotherchris/keepassview/releases/download/v1.2.0/keepassview-darwin-amd64.tar.gz"
      sha256 "8f7f79a730e9f8bdcfa703367f948796411d56e553c1d64b154d20720b183797"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yetanotherchris/keepassview/releases/download/v1.2.0/keepassview-linux-arm64.tar.gz"
      sha256 "df40a075fa59cb5762066172a89c9f835fb862f5ec7747f8a12be6d6c09e0977"
    else
      url "https://github.com/yetanotherchris/keepassview/releases/download/v1.2.0/keepassview-linux-amd64.tar.gz"
      sha256 "cea9b467ff77b9be66e50f48ce3960259e3d4d3a6306c9c5641af775a2bee5c0"
    end
  end

  def install
    bin.install "keepassview-darwin-arm64" => "keepassview" if OS.mac? && Hardware::CPU.arm?
    bin.install "keepassview-darwin-amd64" => "keepassview" if OS.mac? && !Hardware::CPU.arm?
    bin.install "keepassview-linux-arm64" => "keepassview" if OS.linux? && Hardware::CPU.arm?
    bin.install "keepassview-linux-amd64" => "keepassview" if OS.linux? && !Hardware::CPU.arm?
  end

  test do
    assert_match "keepassview 1.2.0", shell_output("#{bin}/keepassview --version")
  end
end
