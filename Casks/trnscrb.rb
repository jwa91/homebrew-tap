cask "trnscrb" do
  version "0.3.0"
  sha256 "3ab569c12160dff0468373abc1247915088766a83514965da40a80e6300b67cc"

  url "https://github.com/jwa91/trnscrb/releases/download/v#{version}/trnscrb-#{version}.dmg"
  name "trnscrb"
  desc "Drop audio, PDFs, and images onto your menu bar to get markdown back"
  homepage "https://github.com/jwa91/trnscrb"

  depends_on macos: ">= :tahoe"

  app "trnscrb.app"
end
