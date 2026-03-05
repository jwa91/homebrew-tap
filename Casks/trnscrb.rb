cask "trnscrb" do
  version "0.1.1"
  sha256 "3ebc9af5414db5c9ca19a20495dca1d17ab35c58fe8f916664d6da961e134344"

  url "https://github.com/jwa91/trnscrb/releases/download/v#{version}/trnscrb-#{version}.dmg"
  name "trnscrb"
  desc "Drop audio, PDFs, and images onto your menu bar to get markdown back"
  homepage "https://github.com/jwa91/trnscrb"

  depends_on macos: ">= :sonoma"

  app "trnscrb.app"

  # Temporary workaround while Apple's notarization queue is still catching up.
  # Remove this block after the notarized DMG has been stapled and re-uploaded.
  postflight do
    system_command "/usr/bin/xattr",
                   args: ["-d", "com.apple.quarantine", "#{appdir}/trnscrb.app"],
                   sudo: false
  end
end
