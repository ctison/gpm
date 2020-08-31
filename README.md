# Github Package Manager (GPM)

> Work in progress. Suggestions, issues and contributions are welcome <3.

GPM is a command line tool to install assets from Github releases.

## Install

```sh
curl -Lo /usr/local/bin/gpm
chmod 500 /usr/local/bin/gpm
```

## Roadmap

- [ ] Logging
- [ ] Configurations
  - [ ] Install location
- [ ] Install command
  - [ ] Maintain a gpm.lock
  - [ ] Install to \$HOME by default and to system with --system
  - [ ] Customize the symbolic link system
  - [ ] Support unpacking .tar .gz .zip .xz
  - [ ] Continue failed downloads
- [ ] Search command
  - [ ] List locally installed assets
  - [ ] List outdated assets
  - [ ] Search by regex or wildcard
  - [ ] Check sums
- [ ] Prune command
- [ ] Term UI
- [ ] Support other platforms
