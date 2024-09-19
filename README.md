# mblog

mblog is a micro blog static site generator written with Go.

## Usage

```
./mblog --tmpl <templates dir>  --in <input dir>  --out <output dir>
```

## Writing

mblog converts markdown posts into html.

### Templates

mblog expects 3 templates:
* index.tmpl is used to generate the index.html
* post.tmpl is used to generate a post
* page.tmpl is used to generate a page

### Index

The index page is the blog's homepage.

### Metada

You can define optional metadata in your pages.

```
title: My title
date: 2024-09-01
template: my_template.tmpl
```
