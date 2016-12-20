(function() {
  'use strict';

  angular
    .module('swan')
    .run(runBlock);

  /** @ngInject */
  function runBlock($log) {

    $log.debug('runBlock end');
  }

})();
