$(document).ready(function() {
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
    $('#uml__modal__' + elem.dataset.umlId).popup({
      opacity: 0.7,
    });
  });
});
