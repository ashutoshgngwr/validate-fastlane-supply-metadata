# Validate Fastlane Supply Metadata

![Docker build](https://github.com/ashutoshgngwr/validate-fastlane-supply-metadata/workflows/Docker/badge.svg)
![Docker image size](https://img.shields.io/docker/image-size/ashutoshgngwr/validate-fastlane-supply-metadata?sort=semver)

A Github Action to statically validate [Fastlane](https://docs.fastlane.tools) metadata
for Android ([supply](https://docs.fastlane.tools/actions/supply/)) using a simple
validation logic written in Golang.

## Features

- Zero config
- Checks title, short description, full description and changelog texts
- Checks promo images
- Checks screenshots
- Tiny docker image ~800KB
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

```yaml
on: [pull_request]
jobs:
  # required to run on Linux because this is a docker container action
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v1
    - uses: ashutoshgngwr/validate-fastlane-supply-metadata@v1
      with:
        fastlaneDir: ./fastlane # optional. default is './fastlane'
```

### Without GitHub actions

The GitHub action relies on a docker image which can be used directly.

```sh
docker run --rm --workdir /app --mount type=bind,source="$(pwd)",target=/app \
   ashutoshgngwr/validate-fastlane-supply-metadata:v1 -help
```

## License

[Apache License 2.0](/LICENSE)
