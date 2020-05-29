# Validate Fastlane Supply Metadata

![Docker](https://github.com/ashutoshgngwr/validate-fastlane-supply-metadata/workflows/Docker/badge.svg)

A Github Action to statically validate Fastlane metadata for Android (supply).

## Example Use Case

In one of my [Android projects](https://github.com/ashutoshgngwr/noice), I was
facing a situation where other developers would translate Fastlane metadata for
the Android app. There was no way to check its validity on the new Pull Requests.
I even created a dedicated CI job to run `fastlane supply` with `validate_only` option.
To run this job for PRs, I would need to expose the service account key
for accessing the Play Store which is a major security flaw.

This action uses a docker image to validate Fastlane's metadata. The docker image
is built from the Go source from this repository. The Go source code is just a
250 lines of validation logic to test all the files in Fastlane metadata against
the constraints from the Play Store listing.

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

## License

[Apache License 2.0](/LICENSE)
