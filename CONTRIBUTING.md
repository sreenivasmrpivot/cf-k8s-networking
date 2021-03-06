# Contributing to CF-K8s-Networking

The Cloud Foundry team uses GitHub and accepts code contributions via [pull
requests](https://help.github.com/articles/about-pull-requests/).

## Prerequisites

Before working on a PR to the CF-K8s-Networking code base, please:

  - reach out to us first via a [GitHub issue](https://github.com/cloudfoundry/cf-k8s-networking/issues),

You can always chat with us on our [Slack #networking channel](https://cloudfoundry.slack.com/app_redirect?channel=CFX13JK7B) ([request an invite](http://slack.cloudfoundry.org/)),

After reaching out to the App Connectivity team and the conclusion is to make a PR, please follow these steps:

1. Ensure that you have either:
   * completed our Contributor License Agreement (CLA) for individuals (if you
     haven't done this already the PR will prompt you to)
   * or, are a [public member](https://help.github.com/articles/publicizing-or-hiding-organization-membership/) of an organization
   that has signed the corporate CLA.
1. Fork the project repository.
1. Create a feature branch (e.g. `git checkout -b good_network`) and make changes on this branch
   * Tests are required for any changes.
1. Push to your fork (e.g. `git push origin good_network`) and [submit a pull request](https://help.github.com/articles/creating-a-pull-request)

Note: All contributions must be sent using GitHub Pull Requests.
We prefer a small, focused pull request with a clear message
that conveys the intent of your change.

## Local development for cf-k8s-networking

### Development dependencies

#### Golang
We currently build with Golang `1.13.x` and use `go mod` for dependencies.

* [`go`](https://golang.org/)
* [`ginkgo`](https://github.com/onsi/ginkgo)

#### k14s Tools
Most of our templating and deploy scripts rely on the [k14s Kubernetes Tools](https://k14s.io/). We recommend installing the latest versions.

* `ytt`
* `kapp`
* `vendir`

#### Integration Tests
Our integration tests spin up temporary
[Kind](https://kubernetes.io/docs/setup/learning-environment/kind/) clusters
using Docker.

* [`docker`](https://docs.docker.com/get-docker/)
* [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start/)


### Running Tests
1. `cd cf-k8s-networking/routecontroller`
2. `ginkgo .`

### Deploying your changes
CF-K8s-Networking is a set of components meant to be integrated into a
[cf-for-k8s](https://github.com/cloudfoundry/cf-for-k8s) deployment.

To deploy your local changes to `cf-k8s-networking` with `cf-for-k8s`, you can
follow these steps:

1. `cd cf-for-k8s`
2. `vendir sync --directory
   config/_ytt_lib/github.com/cloudfoundry/cf-k8s-networking=$PATH_TO_LOCAL_CF_K8S_NETWORKING`
3. Follow docs to install
   [`cf-for-k8s`](https://github.com/cloudfoundry/cf-for-k8s/blob/master/docs/deploy.md)
