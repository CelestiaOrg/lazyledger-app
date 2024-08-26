# Release Guide

The target audience for this guide is maintainers of this repo. In general, the release process is as follows:

1. Create a release candidate
1. Test the release candidate
    1. If the release candidate is not satisfactory, go back to step 1
    1. If the release candidate is satisfactory, create an official release

## Release Candidate

To create a release candidate:

1. Navigate to <https://github.com/celestiaorg/celestia-app/releases/new>.
1. Choose a version tag based on [Semantic Versioning](https://semver.org/). Include the `-rc` suffix.
1. Change the target branch to `v1.x` or `v2.x` based on the version you're releasing.
1. Click **Generate release notes**.
1. Toggle on the **Set as a pre-relase** checkbox.
1. **Publish release**.

After creating the release candidate:

1. Wait until CI passes on the release and verify that prebuilt binaries were attached to the release.
1. Create a PR to bump the celestia-app dependency in [celestia-node](https://github.com/celestiaorg/celestia-node).
1. [Optional] Start a testnet via auto-devops that uses the release candidate. Confirm it works.
1. [Optional] Use the release candidate to sync from genesis. Confirm it works.

## Official Release

Follow the steps above for creating a release candidate:

- The version tag should not include the `-rc` suffix.
- The release notes should contain an **Upgrade Notice** section with notable changes for node operators or library consumers.

After creating the release:

1. Wait until CI passes on the release and verify that prebuilt binaries were attached to the release.
1. Create a PR to bump the celestia-app dependency in [celestia-node](https://github.com/celestiaorg/celestia-node).
