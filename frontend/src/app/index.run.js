(function() {
  'use strict';

  angular
    .module('swan')
    .run(runBlock);

  /** @ngInject */
  function runBlock($rootScope, $stateParams, $state) {
    $rootScope.logProxyBase = BACKEND_URL_BASE.logProxyBase;

    $rootScope.$state = $state;
    $rootScope.$stateParams = $stateParams;
    $rootScope.keys = Object.keys;

  }

})();
