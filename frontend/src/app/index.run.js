(function() {
  'use strict';

  angular
    .module('swan')
    .run(runBlock);

  /** @ngInject */
  function runBlock($log, $rootScope, $stateParams, $state) {

    $rootScope.$state = $state;
    $rootScope.$stateParams = $stateParams;
    $rootScope.keys = Object.keys;

    $log.debug('runBlock end');

  }

})();
