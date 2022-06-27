# drone-email

[![Build Status](https://drone.cryptic.systems/api/badges/volker.raschek/drone-email/status.svg)](https://drone.cryptic.systems/volker.raschek/drone-email)

A Drone CI/CD plugin to send build status notifications via email. The plugin is
currently available for the following architectures:

- x86_64  / amd64
- aarch64 / arm64
- armv7   / arm

## Compile or install the binary locally

Checkout the source code of the project and use `make` to compile or install the
binary locally.

```bash
make all          # compile all targets, including shell completions
make drone-email  # compile only the binary
make install      # install the binary with completions locally
```

## Usage

All params can be defined via cli flags. A list of all provided cli-flags will
be written to `stdout` via `drone-email --help`.

Alternatively can be the flags defined via environment variables or a config file.

### Environment variables

| name                            | description                                     |
| ------------------------------- | ----------------------------------------------- |
| `DRONE_BUILD_CREATED`           | Unix timestamp when the build has been created  |
| `DRONE_BUILD_EVENT`             | Drone event which triggered the build           |
| `DRONE_BUILD_FINISHED`          | Unix timestamp when the build has been finished |
| `DRONE_BUILD_LINK`              | URL to the build pipeline                       |
| `DRONE_BUILD_NUMBER`            | Build number                                    |
| `DRONE_BUILD_STARTED`           | Unix timestamp when the build has been started  |
| `DRONE_BUILD_STATUS`            | Build status                                    |
| `DRONE_COMMIT_AUTHOR_NAME`      | Name of the commit author                       |
| `DRONE_COMMIT_AUTHOR_AVATAR`    | Avatar of the commit author                     |
| `DRONE_COMMIT_AUTHOR_EMAIL`     | EMail of the commit author                      |
| `DRONE_COMMIT_BRANCH`           | Commit branch                                   |
| `DRONE_COMMIT_LINK`             | Link to the commit                              |
| `DRONE_COMMIT_MESSAGE`          | Commit message                                  |
| `DRONE_COMMIT_REF`              | Commit reference                                |
| `DRONE_COMMIT_SHA`              | Commit sha sum                                  |
| `DRONE_DEPLOY_TO`               | Deploy target                                   |
| `DRONE_JOB_EXIT_CODE`           | Job exit code                                   |
| `DRONE_JOB_FINISHED`            | Unix timestamp when the job has been created    |
| `DRONE_JOB_NUMBER`              | Job number                                      |
| `DRONE_JOB_STARTED`             | Unix timestamp when the job has been started    |
| `DRONE_JOB_STATUS`              | Job status                                      |
| `DRONE_PREV_BUILD_NUMBER`       | Previous build number                           |
| `DRONE_PREV_BUILD_STATUS`       | Previous build status                           |
| `DRONE_PREV_COMMIT_SHA`         | Previous commit sha sum                         |
| `DRONE_PULL_REQUEST`            | Number of pull-requests                         |
| `DRONE_REMOTE_URL`              | Clone URL of the repository                     |
| `DRONE_REPO`                    | Name of the repository, including org/owner     |
| `DRONE_REPO_AVATAR`             | Avatar of the repository                        |
| `DRONE_REPO_BRANCH`             | Branch of the repository                        |
| `DRONE_REPO_LINK`               | URL of the repository                           |
| `DRONE_REPO_NAME`               | Name of the repository, without org/owner       |
| `DRONE_REPO_OWNER`              | Org/Owner of the repository                     |
| `DRONE_REPO_PRIVATE`            | Private repository                              |
| `DRONE_REPO_SCM`                | SCM of the repository                           |
| `DRONE_REPO_TRUSTED`            | Trusted repository                              |
| `DRONE_TAG`                     | Tag                                             |
| `DRONE_YAML_SIGNED`             | Yaml is singed                                  |
| `DRONE_YAML_VERIFIED`           | Yaml is trusted                                 |
| `SMTP_FROM_ADDRESS`             | SMTP-From Address                               |
| `SMTP_FROM_NAME`                | SMTP-From Name                                  |
| `SMTP_HELO`                     | SMTP-HELO\EHLO                                  |
| `SMTP_HOST`                     | SMTP-Host                                       |
| `SMTP_MAIL_SUBJECT`             | Overwrite default mail subject template         |
| `SMTP_PASSWORD`                 | SMTP-Password                                   |
| `SMTP_PORT`                     | SMTP-Port                                       |
| `SMTP_START_TLS`                | SMTP-Start-TLS                                  |
| `SMTP_TLS_INSECURE_SKIP_VERIFY` | Trust insecure TLS certificate                  |
| `SMTP_TO_ADDRESSES`             | SMTP-To Addresses                               |
| `SMTP_USERNAME`                 | SMTP-Username                                   |

### Config file

Instead of environment variables, a `config.yaml` can be places in
`/etc/drone-email` or next to the binary.

The yaml should contain the same parameters as the cli flags. For example:

```yaml
drone-build-link: https://drone.example.local/max.mustermann/drone-email/1
drone-build-number: 1
drone-build-status: success
drone-build-started: 1656354006
drone-commit-author-email: max@example.local
drone-commit-author-name: Max Mustermann
drone-commit-branch: master
drone-commit-sha: 06b44cbfa054f146881e7234f1773008f006a756
drone-repo: max.mustermann/drone-email
drone-repo-link: https://git.example.local/max.mustermann/drone-email
smtp-from-address: noreply@example.local
smtp-from-name: noreply
smtp-helo: hostname.example.local
smtp-host: smtp1.example.local
smtp-password: my-password
smtp-username: noreply@example.local
```

## Known issues

### Multiple success emails despite failed ci step

The [drone-runner-kube](https://github.com/drone-runners/drone-runner-kube) does
not define the environment variable `DRONE_PREV_BUILD_STATUS` like the
[drone-runner-docker](https://github.com/drone-runners/drone-runner-docker).
This make it impossible to use the correct email template based on the build
state of the previous step.

Furthermore, the environment variable `DRONE_BUILD_STATUS` is always defined as
`success`, even if the build has failed.

Related issues:

- [Drillster/drone-email](https://github.com/Drillster/drone-email/issues/69)
- [stackoverflow - drone-ci: get status of previous step](https://stackoverflow.com/questions/73096709/drone-ci-get-status-of-previous-step)
