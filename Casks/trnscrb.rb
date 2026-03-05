cask "trnscrb" do
  version "0.1.1"
  sha256 "29c6c574696bfe0f2bf1923b0ebeffbfbd040632f7a6f5e9e94281cd2ac4fb82"

  url "https://github.com/jwa91/trnscrb/releases/download/v#{version}/trnscrb-#{version}.dmg"
  name "trnscrb"
  desc "Drop audio, PDFs, and images onto your menu bar to get markdown back"
  homepage "https://github.com/jwa91/trnscrb"

  depends_on macos: ">= :sonoma"

  app "trnscrb.app"

  # TODO: Remove this postflight block once Apple has accepted our notarization
  # submission. After that, Gatekeeper will recognize the app and the quarantine
  # bypass is no longer needed.
  postflight do
    system_command "/usr/bin/xattr",
                   args: ["-d", "com.apple.quarantine", "#{appdir}/trnscrb.app"],
                   sudo: false
  end
end
