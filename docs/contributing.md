# Contribution guidelines for Prem-Operator

## Welcome

We are excited to have you contribute to the Prem-Operator project!
We welcome all contributions, including bug fixes, feature requests, documentation improvements, and more.

## Making Contributions

### Bug Fixes

- If you've found a bug and have a fix, simply **open a Pull Request (PR)** with your changes.
- Ensure your PR description clearly describes the bug and fix. Include any relevant issue numbers.

### New Features

- While we are open to direct Pull Requests (PRs) for new features, we recommend **opening an issue first** to discuss your ideas with the project maintainers.
- This approach ensures that your efforts are in line with the project's goals and have a higher chance of being accepted.
- However, if you prefer to submit a PR directly, please be aware that it is possible not be reviewed.


## Getting Started

1. **Fork the repository** - Start by forking the project repository to your GitHub account. This creates a personal copy for you to work on.
2. **Clone the forked repository** - Clone your fork to your local machine to start making changes.

```bash
  git clone https://github.com/YOUR_USERNAME/YOUR_FORKED_REPO.git
```
3. **Create a branch** - Create a new branch for the changes you want to make. Use a descriptive branch name.

```bash
  git checkout -b branch-name
```
4. **Make changes** - Make your changes to the codebase.
5. **Commit changes** - Commit your changes with a descriptive commit message.

```bash
  git commit -m 'commit message'
```
6. **Push changes** - Push your changes to your forked repository.

```bash
  git push origin branch-name
```

Before making any changes, please make sure to check out our [Developer Guide](./developer_guide.md) for detailed instructions on code standards, testing procedures, and more to ensure your contributions align with our project's standards.
Check out detailed instructions on [GitHub Workflow Guide.](https://github.com/kubernetes/community/blob/master/contributors/guide/github-workflow.md)

## <a name="signing"></a>Signing your work

The sign-off is a simple line at the end of the explanation for the patch. Your
signature certifies that you wrote the patch or otherwise have the right to pass
it on as an open-source patch. The rules are pretty simple: if you can certify
the below (from [developercertificate.org](http://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe.smith@email.com>

Use your real name (sorry, no pseudonyms or anonymous contributions.)

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`.
