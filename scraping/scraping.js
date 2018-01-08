const Chromy = require('chromy')

let chromy = new Chromy({
  visible: true,
  launchBrowser: false,
});

async function scrape(searchWord) {
  var page = 1;

  while (true) {
    let url = 'https://github.com/search?type=Code&q=' + encodeURIComponent(searchWord) + '&p=' + page;
    console.error('scrape url:' + url);
    await chromy.goto(url);

    let repos = await chromy.evaluate(() => {
      return Array.prototype.map.call(document.getElementsByClassName('code-list-item'), function(e) { return e.getElementsByTagName('a')[2].href; });
    });
    repos.forEach(function(r) {
      console.log(r);
    });

    let hasNext = await chromy.evaluate(() => {
      if (document.getElementsByClassName('next_page').length == 0) {
        return false;
      }
      return document.getElementsByClassName('next_page disabled').length == 0 ? true : false;
    });
    if (!hasNext) {
      console.error('no more page');
      break;
    }

    page++;

    // sleep
    await new Promise(resolve => {
      setTimeout(() => {
        resolve();
      }, 10000);
    });
  }
}

const searchWords = [
  'startuml enduml size:>100 license:mit language:Text',
  'startuml enduml size:>100 license:mit language:Markdown',
  'startuml enduml size:>100 license:mit extension:puml',
  'startuml enduml size:>100 license:mit extension:uml',
  'startuml enduml size:>100 license:mit extension:plantuml',
];

(async function() {
  for (const word of searchWords) {
    await scrape(word);
  }
})();

chromy.close();
