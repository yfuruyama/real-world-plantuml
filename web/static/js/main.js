$(document).ready(function() {
  var isUmlPage = location.pathname.match(/\/umls\/(\d+)/);
  var umlIdInPath = isUmlPage ? isUmlPage[1] : null;

  // setup source copy 
  var clipboard = new Clipboard('.uml__modal__body__source__header__copy');
  clipboard.on('success', function(e) {
    var orig = e.trigger.innerText;
    e.trigger.innerText = "COPIED!";
    setTimeout(function() {
      e.trigger.innerText = orig;
    }, 1000);
  });

  // setup modal
  $('.uml').each(function(i, elem) {
    var umlId = elem.dataset.umlId;
    var autoopen = (umlIdInPath && umlIdInPath == umlId) ? true : false;
    $('#uml__modal__' + umlId).popup({
      autoopen: autoopen,
      opacity: 0.7,
      onopen: function() {
        var originalPath = location.pathname + location.search;
        history.replaceState(originalPath, null, '/umls/' + umlId);
      },
      onclose: function() {
        var originalPath = history.state;
        history.replaceState(null, null, originalPath);
      },
    });
  });
});
