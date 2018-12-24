$(document).ready(function() {
  var isUmlPage = location.pathname.match(/\/umls\/(\d+)/);
  var umlIdInPath = isUmlPage ? isUmlPage[1] : null;
  var allUmlIds = $('.uml').map(function(){
    return this.dataset.umlId
  }).get();

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

  // workaround for https://github.com/yfuruyama/real-world-plantuml/issues/2
  $('.popup_wrapper').each(function(i, elem) {
    elem.style.display = '';
  });

  // keyboard shortcut
  document.onkeydown = function(e) {
    if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
      var isUmlPage = location.pathname.match(/\/umls\/(\d+)/);
      if (!isUmlPage || isUmlPage.length < 2) {
        return;
      }
      var umlId = isUmlPage[1];

      var idx = allUmlIds.indexOf(umlId);
      if (idx !== -1) {
        // hide current
        $('#uml__modal__' + umlId).popup('hide');

        // show previous
        if (e.key === 'ArrowLeft' && idx > 0) {
          $('#uml__modal__' + allUmlIds[idx - 1]).popup('show');
        }

        // show next
        if (e.key === 'ArrowRight' && allUmlIds.length > idx + 1) {
          $('#uml__modal__' + allUmlIds[idx + 1]).popup('show');
        }
      }
    }
  };
});
