real-world-plantuml
===

Source code of https://real-world-plantuml.com/

## Architecture

<img alt="architecture" src="https://github.com/yfuruyama/real-world-plantuml/blob/master/docs/architecture.png" width="700">

## For development

### web

Run server

```
make run GA_TRACKING_ID=${GA_TRACKING_ID}
```

Register dummy UML: access to `/debug/dummy_uml` in your browser

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
