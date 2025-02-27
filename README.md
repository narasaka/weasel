# weasel

small, fast tool to check for broken links in a webpage

## installation

```bash
go install github.com/narasaka/weasel
```

or

```bash
git clone git@github.com/narasaka/weasel
cd weasel
make install
```

## usage

basic (only check links on that page)

```bash
weasel https://example.com
```

recursive check

```bash
weasel -r https://example.com
```

## uninstall

```bash
make uninstall
```
