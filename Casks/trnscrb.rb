cask "trnscrb" do
  version "0.2.0"
  sha256 "321c2e5f9b6d1ce2a224b7a2179a15c34ea33e45ded4d20d78aad1a39b5d546d"

  url "https://github.com/jwa91/trnscrb/releases/download/v#{version}/trnscrb-#{version}.dmg"
  name "trnscrb"
  desc "Drop audio, PDFs, and images onto your menu bar to get markdown back"
  homepage "https://github.com/jwa91/trnscrb"

  depends_on macos: ">= :tahoe"

  app "trnscrb.app"
end
