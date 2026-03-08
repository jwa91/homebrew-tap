cask "trnscrb" do
  version "0.4.1"
  sha256 "58a0d470a3af8a2443736879a7f3d776a3a669682f161f169ada11723427585d"

  url "https://github.com/jwa91/trnscrb/releases/download/v#{version}/trnscrb-#{version}.dmg"
  name "trnscrb"
  desc "Drop audio, PDFs, and images onto your menu bar to get markdown back"
  homepage "https://github.com/jwa91/trnscrb"

  depends_on macos: ">= :tahoe"

  app "trnscrb.app"
end
