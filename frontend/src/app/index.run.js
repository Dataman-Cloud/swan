(function () {
  'use strict';

  angular
    .module('swan')
    .run(runBlock);

  /** @ngInject */
  function runBlock($rootScope, $stateParams, $state) {
    $rootScope.logBase = BACKEND_URL_BASE.logBase;
    $rootScope.monitorBase = BACKEND_URL_BASE.monitorBase;

    $rootScope.$state = $state;
    $rootScope.$stateParams = $stateParams;
    $rootScope.keys = Object.keys;

  }

})();
