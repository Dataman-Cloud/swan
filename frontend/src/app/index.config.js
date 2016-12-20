(function() {
  'use strict';

  angular
    .module('swan')
    .config(config);

  /** @ngInject */
  function config($logProvider, $interpolateProvider, $locationProvider) {
    // Enable log
    $logProvider.debugEnabled(true);

    $locationProvider.html5Mode(true);

    $interpolateProvider.startSymbol('{/');
    $interpolateProvider.endSymbol('/}');
  }

})();
