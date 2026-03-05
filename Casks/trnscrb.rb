cask "trnscrb" do
  version "0.1.0"
  sha256 "fa613c527165647e1988166e6f87a4bf90d8c35d48f0185813bb32b661715bbc"

  url "https://github.com/jwa91/trnscrb/releases/download/v#{version}/trnscrb-#{version}.dmg"
  name "trnscrb"
  desc "Drop audio, PDFs, and images onto your menu bar to get markdown back"
  homepage "https://github.com/jwa91/trnscrb"

  depends_on macos: ">= :sonoma"

  app "trnscrb.app"

  postflight do
    system_command "/usr/bin/xattr",
                   args: ["-d", "com.apple.quarantine", "#{appdir}/trnscrb.app"],
                   sudo: false
  end
end
