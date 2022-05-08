# Validate Fastlane Supply Metadata

[![Docker build](https://github.com/ashutoshgngwr/validate-fastlane-supply-metadata/workflows/Docker/badge.svg)](https://github.com/ashutoshgngwr/validate-fastlane-supply-metadata/actions/workflows/docker.yaml)
[![Docker image size](https://img.shields.io/docker/image-size/ashutoshgngwr/validate-fastlane-supply-metadata?sort=semver)](https://hub.docker.com/r/ashutoshgngwr/validate-fastlane-supply-metadata/tags?page=1&ordering=last_updated)

A Github Action to statically validate [Fastlane](https://docs.fastlane.tools) metadata
for Android ([supply](https://docs.fastlane.tools/actions/supply/)) using a simple
validation logic written in Golang.

## Features

- Zero config
- Supports GitHub file annotations
- Checks title, short description, full description and changelog texts
- Checks promo images
- Checks screenshots
- Optionally checks if a locale is supported by the Play Store Listing
- Tiny docker image ~700KB
- Can be used without GitHub actions

## Example Use Case

In one of my [Android projects](https://github.com/ashutoshgngwr/noice), I was
facing a situation where other developers would translate Fastlane metadata for
the Android app. There was no way to check its validity on the new Pull Requests.
I even created a dedicated CI job to run `fastlane supply` with `validate_only` option.
To run this job for PRs, I would need to expose the service account key
for accessing the Play Store which is a major security flaw.

This action uses a docker image to validate Fastlane's metadata. The docker image
is built from the Go code in this repository. The Go code is ~300 lines of
validation logic to test all the files in Fastlane metadata against the constraints
from the Play Store listing.

## Usage

_See [v1
README](https://github.com/ashutoshgngwr/validate-fastlane-supply-metadata/blob/v1/README.md)
if you're using v1._

```yaml
on: [pull_request]
jobs:
  # required to run on Linux because this is a docker container action
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v1
    - uses: ashutoshgngwr/validate-fastlane-supply-metadata@v2
      with:
        fastlaneDir: ./android-metadata # optional. default is './fastlane/metadata/android'.
        # enable check to validate if a locale is supported by the Play Store Listing.
        usePlayStoreLocales: true # optional. default is false.
```

### Without GitHub actions

The GitHub action uses on a [docker image][dmg] under the hood. You can use it
directly for environments other than GitHub actions.

[dmg]: https://hub.docker.com/r/ashutoshgngwr/validate-fastlane-supply-metadata

```sh
docker run --rm --workdir /app --mount type=bind,source="$(pwd)",target=/app \
   ashutoshgngwr/validate-fastlane-supply-metadata:v2 -help
```

The default entry point accepts the following command-line flags.

```text
-fastlane-path string
    path to the Fastlane Android directory (default "./fastlane/metadata/android")
-ga-file-annotations bool
    enables file annotations for GitHub action (default: false)
-play-store-locales bool
    throw error if a locale isn't recognised by Play Store (default: false)
```

## License

[Apache License 2.0](/LICENSE)
