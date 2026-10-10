#!/usr/bin/env bash
set -euo pipefail

# This prepares artifacts on a trusted Mac. It never publishes them.
[[ "$(uname -s)" == Darwin ]] || { echo 'Release preparation requires macOS.' >&2; exit 1; }
[[ "${MACSWEEP_RELEASE_READY:-}" == true ]] || {
  echo 'Blocked: the documented safety and native Mac release gates are not approved.' >&2
  exit 1
}
: "${RELEASE_TAG:?Set an approved existing tag}"
: "${RELEASE_SHA:?Set its reviewed 40-character commit SHA}"
: "${MACSWEEP_SIGNING_IDENTITY:?Set a Developer ID Application identity}"
: "${MACSWEEP_NOTARY_PROFILE:?Set a notarytool keychain profile}"
[[ "$RELEASE_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Invalid release tag.' >&2; exit 1; }
[[ "$RELEASE_SHA" =~ ^[0-9a-f]{40}$ ]] || { echo 'Invalid commit SHA.' >&2; exit 1; }
[[ "$MACSWEEP_SIGNING_IDENTITY" == 'Developer ID Application: '* ]] || {
  echo 'Ad-hoc signing is not a release signature.' >&2; exit 1;
}
[[ -z "$(git status --porcelain)" ]] || { echo 'Release source must be clean.' >&2; exit 1; }
[[ "$(git rev-parse HEAD)" == "$RELEASE_SHA" ]] || { echo 'Checkout does not match approved SHA.' >&2; exit 1; }
[[ "$(git rev-parse "refs/tags/$RELEASE_TAG^{commit}")" == "$RELEASE_SHA" ]] || {
  echo 'Tag does not match approved SHA.' >&2; exit 1;
}
git merge-base --is-ancestor "$RELEASE_SHA" origin/main
version="$(jq -r .version mygo.json)"
[[ "$RELEASE_TAG" == "v$version" ]] || { echo 'Tag and app version differ.' >&2; exit 1; }
./scripts/check.sh
[[ "${MYGO_INSPECTOR:-0}" != 1 ]] || { echo 'Inspector override is not permitted.' >&2; exit 1; }
CGO_ENABLED=0 go tool mygo build -platform darwin/universal -skip-dmg -skip-notarize -sign "$MACSWEEP_SIGNING_IDENTITY"
app='build/darwin-universal/MacSweep.app'
zip="build/MacSweep-$version.zip"
dmg="build/MacSweep-$version.dmg"
codesign --verify --deep --strict --verbose=2 "$app"
lipo "$app/Contents/MacOS/MacSweep" -verify_arch arm64 x86_64
codesign --display --verbose=4 "$app" 2>&1 | tee build/signature.txt
grep -q '^Authority=Developer ID Application:' build/signature.txt
grep -q 'flags=.*runtime' build/signature.txt
ditto -c -k --sequesterRsrc --keepParent "$app" "$zip"
xcrun notarytool submit "$zip" --keychain-profile "$MACSWEEP_NOTARY_PROFILE" --wait --output-format json > build/notarization-app.json
[[ "$(jq -r .status build/notarization-app.json)" == Accepted ]] || { echo 'App notarization failed.' >&2; exit 1; }
xcrun stapler staple "$app"
xcrun stapler validate "$app"
spctl --assess --type execute --verbose=2 "$app"
# Package the stapled app, not the earlier unstapled submission ZIP.
ditto -c -k --sequesterRsrc --keepParent "$app" "$zip"
image_root="$(mktemp -d)"
trap 'rm -rf "$image_root"' EXIT
ditto "$app" "$image_root/MacSweep.app"
ln -s /Applications "$image_root/Applications"
hdiutil create -volname MacSweep -srcfolder "$image_root" -format UDZO -ov "$dmg"
codesign --sign "$MACSWEEP_SIGNING_IDENTITY" --timestamp "$dmg"
codesign --verify --strict --verbose=2 "$dmg"
xcrun notarytool submit "$dmg" --keychain-profile "$MACSWEEP_NOTARY_PROFILE" --wait --output-format json > build/notarization-dmg.json
[[ "$(jq -r .status build/notarization-dmg.json)" == Accepted ]] || { echo 'DMG notarization failed.' >&2; exit 1; }
xcrun stapler staple "$dmg"
xcrun stapler validate "$dmg"
(
  cd build
  shasum -a 256 "MacSweep-$version.zip" "MacSweep-$version.dmg" > SHA256SUMS
)
printf 'Artifacts prepared for manual review. Nothing was published.\n'
