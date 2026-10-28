(function () {
  function highlightCode(root) {
    var blocks = (root || document).querySelectorAll('pre code');
    for (var i = 0; i < blocks.length; i++) blocks[i].classList.add('microlight');
    if (window.microlight) window.microlight.reset('microlight');
  }
  window.highlightCode = highlightCode;
  highlightCode(document);
})();
