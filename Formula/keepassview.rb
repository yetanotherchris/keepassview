class Keepassview < Formula
  desc "Read-only KeePass (.kdbx) database viewer that runs a local web server"
  homepage "https://github.com/yetanotherchris/keepassview"
  1.0.0 "1.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yetanotherchris/keepassview/releases/download/v1.0.0/keepassview-darwin-arm64.tar.gz"
      sha256 "ae08c234f146022975bce7befae570436cc7d03abff4200568663fddcdc53049"
    else
      url "https://github.com/yetanotherchris/keepassview/releases/download/v1.0.0/keepassview-darwin-amd64.tar.gz"
      sha256 "3da899e699578e2c9348a19527b399e1b1f6c7b9e50183b36b7acf36f080233c"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yetanotherchris/keepassview/releases/download/v1.0.0/keepassview-linux-arm64.tar.gz"
      sha256 "41c3059863336689d65a9dbe00717e03dbb21f8ad1192e8e1204a578c4a3cbe0"
    else
      url "https://github.com/yetanotherchris/keepassview/releases/download/v1.0.0/keepassview-linux-amd64.tar.gz"
      sha256 "5e9d78272aaffa0f476d46170fbf9c75d9357a82c8dfcec7d88ad9daf076602b"
    end
  end

  def install
    bin.install "keepassview-darwin-arm64" => "keepassview" if OS.mac? && Hardware::CPU.arm?
    bin.install "keepassview-darwin-amd64" => "keepassview" if OS.mac? && !Hardware::CPU.arm?
    bin.install "keepassview-linux-arm64" => "keepassview" if OS.linux? && Hardware::CPU.arm?
    bin.install "keepassview-linux-amd64" => "keepassview" if OS.linux? && !Hardware::CPU.arm?
  end

  test do
    assert_match "keepassview 1.0.0", shell_output("#{bin}/keepassview --1.0.0")
  end
end
