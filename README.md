# Google Photos Backup

## Homebrew

```
brew install rupert/repo/google-photos-backup
google-photos-backup --help
```

## GitHub Release

```
VERSION=0.1.0
curl -L -s https://github.com/rupert/google-photos-backup/releases/download/$VERSION/google-photos-backup_linux_amd64 -o google-photos-backup
chmod +x google-photos-backup
./google-photos-backup --help
```

## Compiling from source

1. [Install Go](https://golang.org/doc/install)
2. [Install Dep](https://golang.github.io/dep/docs/installation.html)
3. Download and compile:
   ```
   git clone https://github.com/rupert/google-photos-backup
   cd google-photos-backup
   dep ensure
   go build
   ./google-photos-backup --help
   ```
