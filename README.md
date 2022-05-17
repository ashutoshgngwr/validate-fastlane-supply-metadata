# Validate Fastlane Supply Metadata

[![Docker build](https://github.com/ashutoshgngwr/validate-fastlane-supply-metadata/workflows/Docker/badge.svg)](https://github.com/ashutoshgngwr/validate-fastlane-supply-metadata/actions/workflows/docker.yaml)
[![Docker image size](https://img.shields.io/docker/image-size/ashutoshgngwr/validate-fastlane-supply-metadata?sort=semver)](https://hub.docker.com/r/ashutoshgngwr/validate-fastlane-supply-metadata/tags?page=1&ordering=last_updated)

A Github Action to statically validate [Fastlane][fastlane] metadata for Android
([supply][supply]) using a simple validation logic written in Golang.

[fastlane]: https://docs.fastlane.tools
[supply]: https://docs.fastlane.tools/actions/supply/

## Features

- Zero config
- Supports GitHub file annotations
- Checks title, short description, full description and changelog texts
- Checks promo images
- Checks screenshots
- Optionally checks if Google Play supports provided locales
- Tiny docker image ~700KB
- Usable without GitHub actions

## Example Use Case

In one of my [Android projects][noice], I was facing a situation where other
developers would translate the metadata for the Android app. But, there was no
way to check its validity on a new Pull Request. I could create a dedicated CI
job to run the `fastlane supply` action with the `validate_only` option.
However,to run this job on PRs, I would need to expose a service account key for
accessing the Google Play Developer API. Needless to say that it would be a
security flaw.

This action uses a docker image built using a simple validation logic written in
Go. It validates the Fastlane metadata for Android against the constraints
enforced by Google Play.

[noice]: https://github.com/ashutoshgngwr/noice

## Usage

_See [v1 README][v1-readme] if you're using v1._

[v1-readme]: https://github.com/ashutoshgngwr/validate-fastlane-supply-metadata/blob/v1/README.md

```yaml
on: [pull_request]
jobs:
  # required to run on Linux because this is a docker container action
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v3
    - uses: ashutoshgngwr/validate-fastlane-supply-metadata@v2
      with:
        fastlaneDir: ./android-metadata # optional
        usePlayStoreLocales: true # optional
```

| Option                | Description                                                                                                            |            Default            |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------- | :---------------------------: |
| `fastlaneDir`         | Directory where Fastlane Android metadata is located. It is the directory that contains individual locale directories. | `./fastlane/metadata/android` |
| `usePlayStoreLocales` | Throw an error if Google Play doesn't recognise a locale code. See [available languages][al] on Google Support.        |            `false`            |

[al]: https://support.google.com/googleplay/android-developer/answer/9844778?hl=en#zippy=%2Cview-list-of-available-languages

### Without GitHub actions

The GitHub action runs a [docker image][dmg] under the hood. You can use it
directly for environments other than GitHub actions.

[dmg]: https://hub.docker.com/r/ashutoshgngwr/validate-fastlane-supply-metadata

```sh
docker run --rm --workdir /app --mount type=bind,source="$(pwd)",target=/app \
    ashutoshgngwr/validate-fastlane-supply-metadata:v2 -help
```

The default entry point accepts the following command-line flags.

```txt
-fastlane-path string
    path to the Fastlane Android metadata directory (default "./fastlane/metadata/android")
-ga-file-annotations bool
    enables file annotations for GitHub action (default: false)
-play-store-locales bool
    throw an error if a locale isn't recognised by Google Play (default: false)
```

## License

[Apache License 2.0](/LICENSE)
