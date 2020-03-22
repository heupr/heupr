# heupr

> The service app :wrench:

<img src="https://img.shields.io/badge/solve-problems-blue.svg"> <img src="https://img.shields.io/badge/be-creative-yellow.svg"> <img src="https://img.shields.io/badge/have-fun-red.svg">

[![Build Status](https://travis-ci.org/heupr/heupr.svg?branch=master)](https://travis-ci.org/heupr/heupr) [![Go Report Card](https://goreportcard.com/badge/github.com/heupr/heupr)](https://goreportcard.com/report/github.com/heupr/heupr) [![Coverage Status](https://coveralls.io/repos/github/heupr/heupr/badge.svg?branch=master)](https://coveralls.io/github/heupr/heupr?branch=master)

## Introduction

**Heupr** automates project management for software teams working on GitHub. Our goal is to work towards building features and services that allow developers and managers to stay in the "[flow zone](https://firstround.com/review/track-and-facilitate-your-engineers-flow-states-in-this-simple-way/)" and do what they do best: write code!

Many projects can benefit from automating management tasks and Heupr is designed to provide a platform to do so quickly. The app can be easily installed on a target GitHub repository and configured by including a modular `.heupr.yml` file in the root directory. Any of the core packages, which provide the various feature functionalities such as issue assignment or pull request estimates, can be included in the configuration setup and there are plans to provide inclusion for any third-party packages made available in the community.

See the [open Issues](https://github.com/orgs/heupr/projects/1) for information regarding current limitations to the platform and the status of work being done.

"Heupr" is a portmanteau of the words "heuristic" and "projects" and is pronounced "hew-Per."

## Contributing

Check out our [contribution](https://github.com/heupr/heupr/blob/master/.github/CONTRIBUTING.md) and [conduct](https://github.com/heupr/heupr/blob/master/.github/CODE_OF_CONDUCT.md) guidelines; jump in and get involved!  

We're excited to have you working with us!  

### Code

Here are a few quick points to help get you started working on the core Heupr repository:

- We follow test-driven development (TDD) on this project so be sure to build test cases alongside the production code; [Travis-CI](https://travis-ci.org/heupr/heupr) and [Coveralls](https://coveralls.io/github/heupr/heupr) will ensure that they are run and that our coverage is adequate but feel free to test things out locally too.
- Our overall design goal is to be as "plug-n-play" as possible so new packages or features can be added easily; keep everything _modular_ and _minimal_.
- All of our code should be [clean and readable](https://blog.golang.org/go-fmt-your-code) so be sure to run `gofmt` and `golint` on your code - this is also checked by Travis-CI just to be safe!

### Packages

**NOTE**: External, third-party packages do not yet have support but this is a planned feature. At the moment, if a third-party package is generally beneficial, it could be included in the "built-in" packages provided by Heupr. The guidelines below are for the planned external package support.

If you want to contribute a package that can be used by Heupr's backend, here are some guidelines:

- Packages need to conform to the `Backend` interface provided by the `backend` package in the core repo, [here](https://github.com/heupr/heupr/blob/mini-app/backend/backend.go).
- A publicly accessible `.so` [plugin](https://golang.org/pkg/plugin/) file location needs to be provided via URL so that it can be referenced as a third-party backend package in your or another user's Heupr instance.
- If you feel like you've got a really cool package, feel free to reach out to the project maintainers and request to have it added to the core packages - we'd love to include it!

## Contact

Feel free to reach out to us on [Twitter](https://twitter.com/forstmeier)! if you're unable to find the answers you need in this README, repositories Issues tabs, or the Wiki FAQ page.
