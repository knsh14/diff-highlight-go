# diff-highlight-go
---

git diff-highlight go implementation  
base implementation is [here](https://github.com/git/git/tree/master/contrib/diff-highlight)  
difference is you can't change highlight color. only just inverse.

# Install
```
$ go get github.com/knsh14/diff-highlight-go
```

# Usage
Change pager setting on `.gitconfig`

```
[pager]
    log = diff-highlight-go | less
    show = diff-highlight-go | less
    diff = diff-highlight-go | less
```

