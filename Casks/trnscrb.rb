cask "trnscrb" do
  version "0.4.0"
  sha256 "1b447b640a4a1d935901830f78fa5cbfc69817f533876c35d8ebf17ac87b82ec"

  url "https://github.com/jwa91/trnscrb/releases/download/v#{version}/trnscrb-#{version}.dmg"
  name "trnscrb"
  desc "Drop audio, PDFs, and images onto your menu bar to get markdown back"
  homepage "https://github.com/jwa91/trnscrb"

  depends_on macos: ">= :tahoe"

  app "trnscrb.app"
end
