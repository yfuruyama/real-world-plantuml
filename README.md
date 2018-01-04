real-world-plantuml
===

## For development

### web

Run server

```
make run
```

### indexer

Run server

```
make run GITHUB_API_TOKEN=${GITHUB_API_TOKEN} GCS_BUCKET=${GCS_BUCKET}
```

### renderer

Run server

```
make run
```

### scraping

Launch Chrome with remote debugging enabled

```
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222
```

Run scraping script

```
cd scraping
npm install
node scraping.js > results/YYYYMMDD_01.txt
```
