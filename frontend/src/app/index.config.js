(function() {
  'use strict';

  angular
    .module('swan')
    .config(config);

  /** @ngInject */
  function config($logProvider, $interpolateProvider, $locationProvider, cfpLoadingBarProvider, $httpProvider) {
    // Enable log
    $logProvider.debugEnabled(true);

    $locationProvider.html5Mode(true);

    $interpolateProvider.startSymbol('{/');
    $interpolateProvider.endSymbol('/}');
    $httpProvider.interceptors.push('httpInterceptor');

    cfpLoadingBarProvider.includeSpinner = false;

  }

})();
