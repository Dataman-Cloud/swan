(function () {
  'use strict';

  angular
    .module('swan')
    .controller('AppController', AppController);

  /** @ngInject */
  function AppController(appBackend, $stateParams) {
    var appId = $stateParams.appId;

    var vm = this;
    vm.app = {};

    activate();

    function activate() {
      getAppInfo()
    }

    function getAppInfo() {
      if (appId) {
        appBackend.app({appId: appId}).get(function (data) {
          vm.app = data;
        });
      }
    }
  }
})();
