# Retired: Sift It is now the sift-it cask (brew install jwa91/tap/sift-it).
cask "sift-it-mac" do
  version "0.1.0"
  sha256 "f64b3206e28f07468dfbfa4b54a364db4eba57ea6f47aed47d9d833882c59d8e"

  url "https://github.com/jwa91/sift-it-mac/releases/download/v#{version}/sift-it-mac_#{version}_darwin_arm64.zip"
  name "Sift It"
  desc "One search bar for your files, by what's in them"
  homepage "https://github.com/jwa91/sift-it-mac"

  livecheck do
    skip "Auto-generated on release."
  end

  disable! date: "2026-10-24", because: "is now part of the sift-it cask", replacement_cask: "sift-it"

  depends_on arch: :arm64
  depends_on cask: "jwa91/tap/sift-it"
  depends_on macos: :sonoma

  app "Sift It.app"

  uninstall quit: "dev.sift-it.mac"

  zap trash: "~/Library/Preferences/dev.sift-it.mac.plist"

  caveats <<~EOS
    Sift It searches what sift-it indexes. Set that up once, if you haven't:
      sift-it model download
      sift-it folders add ~/Documents
      sift-it service install
    Then open Sift It: Control-Shift-Space opens the search from any app.
  EOS
end
