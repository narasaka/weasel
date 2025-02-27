# weasel

small, fast tool to check for broken links in a webpage. i made this because i
needed it for a project. it might come in handy for you too.

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

> [!WARNING]\
> due to the concurrent nature of the script when checking the links, you might
> often get a 429 error (too many requests) when doing this.

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
